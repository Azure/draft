# remove previous tests
echo "Removing previous integration configs"
rm -rf ./integration/*
echo "Removing previous integration workflows"
rm ../.github/workflows/draftv2-integrations.yml

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
          if-no-files-found: error" > ../.github/workflows/draftv2-integrations.yml


# read config and add integration test for each language
cat integration_config.json | jq -c '.[]' | while read -r test; 
do 
    lang=$(echo $test | jq '.language' -r)
    port=$(echo $test | jq '.port' -r)
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
done
