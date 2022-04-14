# remove previous tests
echo "Removing previous integration configs"
rm -rf ./integration/*
echo "Removing previous integration workflows"
rm ../.github/workflows/integration-linux.yml
rm ../.github/workflows/integration-windows.yml

# add build to workflow
echo "name: DraftV2 Linux Integrations

on:
  pull_request_review:
    types: [submitted]
  workflow_dispatch:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.18
      - name: make
        run: make
      - uses: actions/upload-artifact@v2
        with:
          name: helm-skaffold
          path: ./test/skaffold.yaml
          if-no-files-found: error
      - uses: actions/upload-artifact@v2
        with:
          name: draftv2-binary
          path: ./draftv2
          if-no-files-found: error" > ../.github/workflows/integration-linux.yml

echo "name: DraftV2 Windows Integrations

on:
  pull_request_review:
    types: [submitted]
  workflow_dispatch:

jobs:
  build:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.18
      - name: make
        run: make
      - uses: actions/upload-artifact@v2
        with:
          name: draftv2-binary
          path: ./draftv2.exe
          if-no-files-found: error
      - uses: actions/upload-artifact@v2
        with:
          name: check_windows_helm
          path: ./test/check_windows_helm.ps1
          if-no-files-found: error
      - uses: actions/upload-artifact@v2
        with:
          name: check_windows_kustomize
          path: ./test/check_windows_kustomize.ps1
          if-no-files-found: error" > ../.github/workflows/integration-windows.yml



# read config and add integration test for each language
cat integration_config.json | jq -c '.[]' | while read -r test; 
do 
    # extract from json
    lang=$(echo $test | jq '.language' -r)
    port=$(echo $test | jq '.port' -r)
    repo=$(echo $test | jq '.repo' -r)
    echo "Adding $lang with port $port"

    mkdir ./integration/$lang

    # create helm.yaml
    echo "deployType: \"Helm\"
languageType: \"$lang\"
deployVariables:
  - name: \"PORT\"
    value: \"$port\"
  - name: \"APPNAME\"
    value: \"testapp\"
languageVariables:
  - name: \"PORT\"
    value: \"$port\"" > ./integration/$lang/helm.yaml

    # create kustomize.yaml
    echo "deployType: \"kustomize\"
languageType: \"$lang\"
deployVariables:
  - name: \"PORT\"
    value: \"$port\"
languageVariables:
  - name: \"PORT\"
    value: \"$port\"" > ./integration/$lang/kustomize.yaml

    # create kustomize.yaml
    echo "deployType: \"kustomize\"
languageType: \"$lang\"
deployVariables:
  - name: \"PORT\"
    value: \"$port\"
  - name: \"APPNAME\"
    value: \"testapp\"
languageVariables:
  - name: \"PORT\"
    value: \"$port\"" > ./integration/$lang/manifest.yaml

    # create helm workflow
    echo "
  $lang-helm:
    runs-on: ubuntu-latest
    services:
      registry:
        image: registry:2
        ports:
          - 5000:5000
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draftv2-binary
      - run: chmod +x ./draftv2
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: rm -rf ./langtest/manifests && rm -f ./langtest/Dockerfile ./langtest/.dockerignore
      - run: ./draftv2 -v create -c ./test/integration/$lang/helm.yaml -d ./langtest/
      - run: ./draftv2 -v generate-workflow -d ./langtest/ -c someAksCluster -r someRegistry -g someResourceGroup --container-name someContainer
      # Runs Helm to create manifest files
      - name: Bake deployment
        uses: azure/k8s-bake@v2.1
        with:
          renderEngine: 'helm'
          helmChart: ./langtest/charts
          overrideFiles: ./langtest/charts/values.yaml
          overrides: |
            replicas:2
          helm-version: 'latest'
        id: bake
      - name: Set up QEMU
        uses: docker/setup-qemu-action@v1
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v1
        with:
          driver-opts: network=host
      - name: Build and push
        uses: docker/build-push-action@v2
        with:
          context: ./langtest/
          push: true
          tags: localhost:5000/name/testapp:latest
      # Deploys application based on manifest files from previous step
      - name: Deploy application
        uses: Azure/k8s-deploy@v3.0
        with:
          action: deploy
          manifests: \${{ steps.bake.outputs.manifestsBundle }}
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - run: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
      - run: cd ./langtest && skaffold run" >> ../.github/workflows/integration-linux.yml

    # create kustomize workflow
    echo "
  $lang-kustomize:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draftv2-binary
      - run: chmod +x ./draftv2
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: rm -rf ./langtest/manifests && rm -f ./langtest/Dockerfile ./langtest/.dockerignore
      - run: ./draftv2 -v create -c ./test/integration/$lang/kustomize.yaml -d ./langtest/
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - run: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
      - run: cd ./langtest && skaffold init --force
      - run: cd ./langtest && skaffold run" >> ../.github/workflows/integration-linux.yml

  # create manifests workflow
    echo "
  $lang-manifests:
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draftv2-binary
      - run: chmod +x ./draftv2
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: rm -rf ./langtest/manifests && rm -f ./langtest/Dockerfile ./langtest/.dockerignore
      - run: ./draftv2 -v create -c ./test/integration/$lang/manifest.yaml -d ./langtest/
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - run: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
      - run: cd ./langtest && skaffold init --force
      - run: cd ./langtest && skaffold run" >> ../.github/workflows/integration-linux.yml

    # create helm workflow
    echo "
  $lang-helm:
    runs-on: windows-latest
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draftv2-binary
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: Remove-Item ./langtest/manifests -Recurse -Force -ErrorAction Ignore
      - run: Remove-Item ./langtest/Dockerfile -ErrorAction Ignore
      - run: Remove-Item ./langtest/.dockerignore -ErrorAction Ignore
      - run: ./draftv2.exe -v create -c ./test/integration/$lang/helm.yaml -d ./langtest/
      - uses: actions/download-artifact@v2
        with:
          name: check_windows_helm
          path: ./langtest/
      - run: ./check_windows_helm.ps1
        working-directory: ./langtest/" >> ../.github/workflows/integration-windows.yml

    # create kustomize workflow
    echo "
  $lang-kustomize:
    runs-on: windows-latest
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draftv2-binary
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: Remove-Item ./langtest/manifests -Recurse -Force -ErrorAction Ignore
      - run: Remove-Item ./langtest/Dockerfile -ErrorAction Ignore
      - run: Remove-Item ./langtest/.dockerignore -ErrorAction Ignore
      - run: ./draftv2.exe -v create -c ./test/integration/$lang/kustomize.yaml -d ./langtest/
      - uses: actions/download-artifact@v2
        with:
          name: check_windows_kustomize
          path: ./langtest/
      - run: ./check_windows_kustomize.ps1
        working-directory: ./langtest/" >> ../.github/workflows/integration-windows.yml
done
