# Change Log

## [0.0.29] - 2023-02-16

### Fixed
- Bug fix for helm deployments where namespace was created but not respected at the deployment level

## [0.0.28] - 2023-02-13

**BREAKING** changes to `IMAGE` variable

### Added
- New, optional `disablePrompt` property on Builder Variables in draft config (#180):
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
- For all draft substitutions, draft will now error if unsubstituted variables are found in the final output (#175)

### Changed
- **BREAKING** the `IMAGE` variable no longer can include an image tag. The `TAG` variable should be used instead (#176)
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