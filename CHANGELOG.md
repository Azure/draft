# Change Log

## [0.0.34] - 2024-05-16

### Added

- [#277](https://github.com/Azure/draft/pull/277) Updates/Fixes for draft. Addes generator label to manifests
- [#275](https://github.com/Azure/draft/pull/275) Add NOTICE file
- [#274](https://github.com/Azure/draft/pull/274) gomodule multistage build
- [#273](https://github.com/Azure/draft/pull/273) finighing safeguard additions
- [#272](https://github.com/Azure/draft/pull/272) sdk calls for assignsprole
- [#271](https://github.com/Azure/draft/pull/271) update draft to go 1.22
- [#269](https://github.com/Azure/draft/pull/269) yaml file extension validation
- [#262](https://github.com/Azure/draft/pull/262) sdk calls for getTenantID
- [#242](https://github.com/Azure/draft/pull/242) changes in correlation with new GH action permission changes


## [0.0.33] - 2023-08-07

### Added

- [#220](https://github.com/Azure/draft/pull/220) Update readme for supported flags
- [#219](https://github.com/Azure/draft/pull/219) Retry releases to get tag name
- [#218](https://github.com/Azure/draft/pull/218) Add gradle wrapper detection
- [#217](https://github.com/Azure/draft/pull/217) Add python entrypoint detection
- [#215](https://github.com/Azure/draft/pull/215) Add reporeader interface and an example extractor for python
- [#213](https://github.com/Azure/draft/pull/213) Integration test for multiple OS

### Fixed

- [#225](https://github.com/Azure/draft/pull/225) Fix variable substitution in `generate-workflow`
- [#216](https://github.com/Azure/draft/pull/216) bump rust version to fix e2e

## [0.0.32] - 2023-04-10

### Added

- [#197](https://github.com/Azure/draft/pull/197) Add dry run support to `update` command
- [#191](https://github.com/Azure/draft/pull/191) Add variable flag to `create` command

### Fixed

- [#196](https://github.com/Azure/draft/pull/196) Update deprecated node12 actions
- [#207](https://github.com/Azure/draft/pull/207) Default deploy variable fixed

### Changed 

- [#194](https://github.com/Azure/draft/pull/194) Move generate workflow templates to embedded file system

## [0.0.31] - 2023-03-14

### Added

- [#189](https://github.com/Azure/draft/pull/189) Add `example` package to show consumption patterns

### Fixed

- [#190](https://github.com/Azure/draft/pull/190) Add integration test and Dockerfile fix for `go` language without modules

## [0.0.30] - 2023-02-24

### Changed

- [#187](https://github.com/Azure/draft/pull/187) OpenJDK Docker image has been deprecated and doesnt have JRE images for  Java 11+. This change moves our Java images to Eclipse-Temurin.

## [0.0.29] - 2023-02-16

### Fixed
- [#183](https://github.com/Azure/draft/pull/183) Bug fix for helm deployments where namespace was created but not respected at the deployment level

## [0.0.28] - 2023-02-13

**BREAKING** changes to `IMAGE` variable

### Added
- New, optional `disablePrompt` property on Builder Variables in draft config [#180](https://github.com/Azure/draft/pull/180):
  - Default Value: `false` 
  - Variables with `disablePrompt: true` will not be prompted for when running `draft interactive commands`
  - Variables with `disablePrompt: true` can still be supplied via flags (`draft create --var TAG=latest`) or draft config files
  - Example Usage:
    ```
      # draft.yaml
      variables:
      - name: "TAG"
        description: "the tag of the image to be built"
        disablePrompt: true #  New optional field that is used to disable the prompt for this variable
      ...
      ```
- For all draft substitutions, draft will now error if unsubstituted variables are found in the final output [#175](https://github.com/Azure/draft/pull/175)

### Changed
- **BREAKING** the `IMAGE` variable no longer can include an image tag. The `TAG` variable should be used instead [#176](https://github.com/Azure/draft/pull/176)
- **BREAKING** the `imageKey` variable on the `helm` deployment type has been renamed to `image` to be consistent with the supplied starter workflows (#176)
  - Re-running `draft create` will update existing files to follow the new convention

## [0.0.27] - 2022-12-9

### Added
- New `displayName` and `variables.exampeValues` properties in draft.yaml
  ```yaml
    # draft.yaml
    language: swift
    displayName: Swift # Add a display name for the selected resource (language/deploymentType/addon)
    variables:
    - name: "VERSION"
      description: "the version of swift used by the application"
      exampleValues: ["5.5","5.4"] #  New optional field that is used to populate draft info, and which could be used in the cli for suggestions in the future.
      variableDefaults:
    - name: "VERSION"
      value: 5.5
    ```
- Added `--dry-run` and `--dry-run-file` flags to `create` command
  - `--dry-run` enables dry run mode in which no files are written to disk, prints the dry run summary to stdout
  - `--dry-run-file` specifies a file to write the dry run summary in json format (requires `--dry-run` flag)
  ```json
  # Example dry run output
  {
    "variables": {
      "APPNAME": "testapp",
      "BUILDERVERSION": "null",
      "IMAGENAME": "testapp",
      "LANGUAGE": "gomodule",  # Note that this variable is in addition to the draft config variables
      "NAMESPACE": "default",
      "PORT": "1323",
      "SERVICEPORT": "80",
      "VERSION": "1.18"
    },
    "filesToWrite": [
      "langtest/.dockerignore",
      "langtest/Dockerfile",
      "langtest/charts/.helmignore",
      "langtest/charts/Chart.yaml",
      "langtest/charts/production.yaml",
      "langtest/charts/templates/_helpers.tpl",
      "langtest/charts/templates/deployment.yaml",
      "langtest/charts/templates/namespace.yaml",
      "langtest/charts/templates/service.yaml",
      "langtest/charts/values.yaml"
    ]
  }
  ```
  
### Changed

- **BREAKING** - `info` command output format includes additional information for supported languages with the following format changes:
    - keys are now camelCase
    - `supportedLanguages` is now an array of objects, enriched with `displayName` and `exampleValues`
    ```json
    # Example of the new info command output format
    {
      # keys are now camelCase, so supported_languages is now supportedLanguages
      "supportedLanguages": [
        {
          "name": "javascript",
          "displayName": "JavaScript", # new field
          "exampleValues": {           # new field
            "VERSION": [
              "14.0",
              "16.0"
            ]
          }
        },
        ...
      ],
     # keys are now camelCase, so supported_deployment_types is now supportedDeploymentTypes
     "supportedDeploymentTypes": [
        "helm",
        ...
      ]
    }
    ```
  
## [0.0.26] - 2022-11-16

### Added

- The new `draft info` command from #157 prints supported language and field information in json format.
- An integration test was added for the installation shell script to better ensure that the script works as expected.

### Fixed

- File path output for root locations had a bug with string-formatted paths. The `path.Join` method has been substituted to fix this.

### Changed

- Remaining uses of the `viper` library have been migrated to `gopkg.in/yaml.v3` for consistency.
- Unused files in the `web` package have been removed.
- Minor reorganization across the `config` and `addons` packages to reduce the number of exported functions and types.
