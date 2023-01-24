# remove previous tests
echo "Removing previous integration configs"
rm -rf ./integration/*
echo "Removing previous integration workflows"
rm ../.github/workflows/integration-linux.yml
rm ../.github/workflows/integration-windows.yml

# add build to workflow
echo "name: draft Linux Integrations

on:
  push:
    branches: [ int-tests ]
  workflow_dispatch:

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.18.2
      - name: make
        run: make
      - uses: actions/upload-artifact@v2
        with:
          name: helm-skaffold
          path: ./test/skaffold.yaml
          if-no-files-found: error
      - uses: actions/upload-artifact@v2
        with:
          name: draft-binary
          path: ./draft
          if-no-files-found: error" > ../.github/workflows/integration-linux.yml
# read config and add integration test for each language
cat integration_config.json | jq -c '.[]' | while read -r test;
do
    note="# this file is generated using gen_integration.sh"
    # extract from json
    lang=$(echo $test | jq '.language' -r)
    version=$(echo $test | jq '.version' -r)
    builderversion=$(echo $test | jq '.builderversion' -r)
    port=$(echo $test | jq '.port' -r)
    serviceport=$(echo $test | jq '.serviceport' -r)
    repo=$(echo $test | jq '.repo' -r)
    # addon integration testing vars
    ingress_test_args="-a webapp_routing --variable ingress-tls-cert-keyvault-uri=test.cert.keyvault.uri --variable ingress-use-osm-mtls=true --variable ingress-host=host1"
    subf="subfolder"
    echo "Adding $lang with port $port"

    mkdir ./integration/$lang

    # create helm.yaml
    echo "$note
deployType: \"Helm\"
languageType: \"$lang\"
deployVariables:
  - name: \"PORT\"
    value: \"$port\"
  - name: \"SERVICEPORT\"
    value: \"$serviceport\"
  - name: \"APPNAME\"
    value: \"testapp\"
languageVariables:
  - name: \"VERSION\"
    value: \"$version\"
  - name: \"BUILDERVERSION\"
    value: \"$builderversion\"
  - name: \"PORT\"
    value: \"$port\"" > ./integration/$lang/helm.yaml

    # create helm workflow
    echo "
  $lang-helm-dry-run:
      runs-on: ubuntu-latest
      needs: build
      steps:
        - uses: actions/checkout@v2
        - uses: actions/download-artifact@v2
          with:
            name: draft-binary
        - run: chmod +x ./draft
        - run: mkdir ./langtest
        - uses: actions/checkout@v2
          with:
            repository: $repo
            path: ./langtest
        - name: Execute Dry Run
          run: |
            mkdir -p test/temp
            ./draft --dry-run --dry-run-file test/temp/dry-run.json \
            create -c ./test/integration/$lang/helm.yaml \
            -d ./langtest/ --skip-file-detection
        - name: Validate JSON
          run: |
            npm install -g ajv-cli@5.0.0
            ajv validate -s test/dry_run_schema.json -d test/temp/dry-run.json
  $lang-helm-create-update:
    runs-on: ubuntu-latest
    services:
      registry:
        image: registry:2
        ports:
          - 5000:5000
    needs: $lang-helm-dry-run
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draft-binary
      - run: chmod +x ./draft
      - run: mkdir ./langtest
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest
      - run: rm -rf ./langtest/manifests && rm -f ./langtest/Dockerfile ./langtest/.dockerignore
      - run: ./draft -v create -c ./test/integration/$lang/helm.yaml -d ./langtest/
      - run: ./draft -b main -v generate-workflow -d ./langtest/ -c someAksCluster -r someRegistry -g someResourceGroup --container-name someContainer
      - run: ./draft -v update -d ./langtest/ $ingress_test_args
      - name: start minikube
        id: minikube
        uses: medyagh/setup-minikube@master
      - name: Build image
        run: |
          export SHELL=/bin/bash
          eval \$(minikube -p minikube docker-env)
          docker build -f ./langtest/Dockerfile -t testapp ./langtest/
          echo -n "verifying images:"
          docker images
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
      # Deploys application based on manifest files from previous step
      - name: Deploy application
        uses: Azure/k8s-deploy@v3.0
        continue-on-error: true
        id: deploy
        with:
          action: deploy
          manifests: \${{ steps.bake.outputs.manifestsBundle }}
      - name: Check default namespace
        if: steps.deploy.outcome != 'success'
        run: kubectl get po
      - name: Fail if any error
        if: steps.deploy.outcome != 'success'
        run: exit 6" >> ../.github/workflows/integration-linux.yml
    # create helm workflow
    echo "
  $lang-helm-create:
    runs-on: windows-latest
    needs: build
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draft-binary
      - run: mkdir -p ./langtest/$subf
      - uses: actions/checkout@v2
        with:
          repository: $repo
          path: ./langtest/$subf
      - run: Remove-Item ./langtest/$subf/manifests -Recurse -Force -ErrorAction Ignore
      - run: Remove-Item ./langtest/$subf/Dockerfile -ErrorAction Ignore
      - run: Remove-Item ./langtest/$subf/.dockerignore -ErrorAction Ignore
      - run: ./draft.exe -v create -c ./test/integration/$lang/helm.yaml -d ./langtest/ -s $subf
      - uses: actions/download-artifact@v2
        with:
          name: check_windows_helm
          path: ./langtest/$subf
      - run: ./check_windows_helm.ps1
        working-directory: ./langtest/$subf
      - uses: actions/upload-artifact@v3
        with:
          name: $lang-helm-create
          path: ./langtest/$subf
  $lang-helm-update:
    needs: $lang-helm-create
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/download-artifact@v2
        with:
          name: draft-binary
      - uses: actions/download-artifact@v3
        with:
          name: $lang-helm-create
          path: ./langtest/$subf
      - run: Remove-Item ./langtest/$subf/charts/templates/ingress.yaml -Recurse -Force -ErrorAction Ignore
      - run: ./draft.exe -v update -d ./langtest/ -s $subf $ingress_test_args
      - uses: actions/download-artifact@v2
        with:
          name: check_windows_addon_helm
          path: ./langtest/$subf
      - run: ./check_windows_addon_helm.ps1
        working-directory: ./langtest/$subf" >> ../.github/workflows/integration-windows.yml
done
