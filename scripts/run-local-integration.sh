#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  echo "Usage: $0 <deploy-type> <language> [repo]"
  echo "  deploy-type: helm | kustomize | manifest"
  echo "  language: integration language folder (e.g. swift, go, python)"
  echo "  repo: optional git repo (defaults to integration-tests.yml value)"
  exit 1
fi

deploy_type="$1"
language="$2"
repo="${3:-}"

case "$deploy_type" in
  helm|kustomize|manifest) ;;
  *)
    echo "Unsupported deploy type: $deploy_type"
    exit 1
    ;;
esac

if [ ! -x ./draft ]; then
  echo "./draft not found or not executable. Build it first (e.g. make)."
  exit 1
fi

if [ -z "$repo" ]; then
  repo=$(awk -v lang="$language" '
    $1 == "language:" { gsub(/"/, "", $2); if ($2 == lang) { found=1 } }
    found && $1 == "repo:" { gsub(/"/, "", $2); print $2; exit }
  ' .github/workflows/integration-tests.yml)
fi

if [ -z "$repo" ]; then
  echo "Repo not provided and not found in integration-tests.yml for language '$language'."
  exit 1
fi

echo "==> Using repo: $repo"
cache_root="${DRAFT_LOCAL_CACHE:-$HOME/.cache/draft}"
cache_dir="$cache_root/repos"
repo_cache_dir="$cache_dir/${repo//\//_}"
repo_dir="$repo_cache_dir/repo"
work_dir="$repo_cache_dir"

mkdir -p "$cache_dir"

if [ ! -d "$repo_dir/.git" ]; then
  echo "==> Cloning into cache: $repo_cache_dir"
  git clone "https://github.com/$repo.git" "$repo_dir"
else
  echo "==> Using cached repo: $repo_cache_dir"
fi

git -C "$repo_dir" fetch --quiet origin
git -C "$repo_dir" reset --hard origin/HEAD

commit_sha=$(git -C "$repo_dir" rev-parse HEAD)
echo "==> Using commit: $commit_sha"

config_file="./test/integration/$language/$deploy_type.yaml"
if ! command -v yq >/dev/null 2>&1; then
  echo "yq is required to read APPNAME from $config_file" >&2
  exit 1
fi
appname=$(yq e '.deployVariables[] | select(.name == "APPNAME") | .value' "$config_file")
if [ -z "$appname" ] || [ "$appname" = "null" ]; then
  appname=$(yq e '.languageVariables[] | select(.name == "APPNAME") | .value' "$config_file")
fi
if [ -z "$appname" ] || [ "$appname" = "null" ]; then
  echo "APPNAME not found in $config_file" >&2
  exit 1
fi
product_name="$appname"
echo "==> Using APPNAME from config: $appname"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind not found in PATH. Install kind first."
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not running. Start Docker Desktop and retry."
  exit 1
fi

registry_port=5000
registry_name="kind-registry"
cluster_name="kind"
if ! kind get clusters | grep -q "^$cluster_name$"; then
  kind create cluster --name "$cluster_name"
fi

if ! docker ps --format '{{.Names}}' | grep -q "^$registry_name$"; then
  docker run -d --restart=always -p "$registry_port:5000" --name "$registry_name" registry:2
fi

if ! docker network inspect kind >/dev/null 2>&1; then
  echo "kind docker network not found."
  exit 1
fi

if ! docker network inspect kind | grep -q "$registry_name"; then
  docker network connect kind "$registry_name" || true
fi

push_registry="localhost:$registry_port"
pull_registry="$registry_name:$registry_port"
cat >"$work_dir/local-registry" <<EOF
$pull_registry
EOF

echo "==> Configuring kind to use local registry"
registry_dir="/etc/containerd/certs.d/$pull_registry"
cat <<EOF | docker exec -i "$cluster_name-control-plane" sh -c "mkdir -p $registry_dir && cat > $registry_dir/hosts.toml"
server = "http://$pull_registry"
[host."http://$pull_registry"]
  capabilities = ["pull", "resolve", "push"]
EOF
docker exec "$cluster_name-control-plane" sh -c "systemctl restart containerd" >/dev/null 2>&1 \
  || docker exec "$cluster_name-control-plane" sh -c "kill -SIGHUP $(pidof containerd)" >/dev/null 2>&1 || true

image="$push_registry/$appname"

echo "==> Generating manifests"
target_dir="$repo_dir"

rm -rf "$target_dir/manifests" "$target_dir/charts" "$target_dir/base" "$target_dir/overlays"
rm -f "$target_dir/Dockerfile" "$target_dir/.dockerignore"

extra_vars=()

case "$deploy_type" in
  helm)
    ./draft -v create -c "./test/integration/$language/helm.yaml" -d "$target_dir" --skip-file-detection ${extra_vars:+"${extra_vars[@]}"}
    ;;
  kustomize)
    ./draft -v create -c "./test/integration/$language/kustomize.yaml" -d "$target_dir" ${extra_vars:+"${extra_vars[@]}"}
    ;;
  manifest)
    ./draft -v create -c "./test/integration/$language/manifest.yaml" -d "$target_dir" --skip-file-detection ${extra_vars:+"${extra_vars[@]}"}
    ;;
esac


echo "==> Building and pushing image: $image"
if docker buildx inspect draft-builder >/dev/null 2>&1; then
  docker buildx use draft-builder
else
  docker buildx create --use --name draft-builder --driver docker-container --driver-opt network=host
fi
docker buildx inspect --bootstrap
docker buildx build -f "$target_dir/Dockerfile" -t "$image" --push "$target_dir"

echo "==> Deploying"
case "$deploy_type" in
  helm)
    helm upgrade --install test-release "$target_dir/charts" \
      --set image.repository="$pull_registry/$appname" \
      --set image.tag="latest" \
      --set replicas=1
    kubectl rollout status deployment/test-release-$appname --timeout=5m
    kubectl port-forward svc/test-release-$appname 18080:80 >/tmp/port-forward.log 2>&1 &
    port_forward_pid=$!
    ;;
  kustomize)
    kubectl apply -k "$target_dir/base"
    kubectl set image deployment/$appname $appname="$image" --record
    kubectl rollout status deployment/$appname --timeout=5m
    kubectl port-forward svc/$appname 18080:80 >/tmp/port-forward.log 2>&1 &
    port_forward_pid=$!
    ;;
  manifest)
    kubectl apply -f "$target_dir/manifests"
    kubectl set image deployment/$appname $appname="$image" --record
    kubectl rollout status deployment/$appname --timeout=5m
    kubectl port-forward svc/$appname 18080:80 >/tmp/port-forward.log 2>&1 &
    port_forward_pid=$!
    ;;
esac

sleep 3
echo "==> Curling service"
set +e
curl -fsSL http://localhost:18080
curl_status=$?
set -e

if [ -n "${port_forward_pid:-}" ]; then
  kill "$port_forward_pid" >/dev/null 2>&1 || true
fi

if [ "$curl_status" -ne 0 ]; then
  echo "curl failed (exit $curl_status)."
  exit "$curl_status"
fi

echo "OK"
