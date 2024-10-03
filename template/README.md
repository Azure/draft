# TEMPLATE DEFINITION DOCS

## Definitions

All templates are defined within the `./template` directory with a cluster of go template files accompanied by a `draft.yaml` file.

### draft.yaml

The `draft.yaml` file contains the metadata needed to define a Template in Draft. The structure of the `draft.yaml` is as follows:

- `templateName` - The name of the template
- `type` - The type of template
- `description` - Description of template contents/functionality
- `versions` - the range/list of version definitions for this template
- `defaultVersions` - If no version is passed to a template this will be used
- `parameters` - a struct containing information on each parameter to the template
  - `name` - the parameter name associated to the gotemplate variable
  - `description` - description of what the parameter is used for
  - `type` - defines the type of the parameter
  - `kind` - defines the kind of parameter, useful for prompting and validation within portal/cli/vsce
  - `required` - defines if the parameter is required for the template
  - `default` - struct containing information on specific parameters default value
    - `value` - the parameters default value
    - `referenceVar` - the variable to reference if one is not provided
  - `versions` - the versions this item is used for

For the `type` parameters at the template level we currently have 4 definitions:
- `deployment` - the base k8s deployment + service + namespace
- `dockerfile` - representing a dockerfile for a specific language
- `workflow` - representing a GitHub Action, ADO Pipeline, or similar
- `manifest` - a generic k8s manifest. Think PDB, Ingress, HPA that can be added to an existing `deployment`

For the `type` parameter at the variable level, this is in line with structured types: `int`, `float`, `string`, `bool`, `object`.

For the `kind` parameter, this will be used for validation and transformation logic on the input. As an example, `azureResourceGroup` and `azureResourceName` can be validated as defined.

### Validation

Within the [draft config teamplate tests](../pkg/config/draftconfig_template_test.go) there is validation logic to make sure all `draft.yaml` definitions adhere to:
- Unique `templateName`'s
- Valid Template `type`'s
- Valid parameter `type`'s
- Valid parameter `kind`'s