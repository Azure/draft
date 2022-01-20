# remove previous tests
echo "Removing previous integration configs"
rm -rf ./integration/*
echo "Removing previous integration workflows"
rm ../.github/workflows/integration.yml

# add build to workflow
echo "name: DraftV2 Integrations

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.17
      - name: make
        run: make
      - uses: actions/upload-artifact@v2
        with:
          name: draftv2-binary
          path: ./draftv2
          if-no-files-found: error" > ../.github/workflows/integration.yml


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

    # create helm workflow
    echo "
  $lang-helm:
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
      - run: ./draftv2 -v create -c ./test/integration/$lang/helm.yaml ./langtest/
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - run: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
      - run: cd ./langtest && skaffold run" >> ../.github/workflows/integration.yml

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
      - run: ./draftv2 -v create -c ./test/integration/$lang/kustomize.yaml ./langtest/
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - run: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/
      - run: cd ./langtest && skaffold init --force
      - run: cd ./langtest && skaffold run" >> ../.github/workflows/integration.yml
done
