package config

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

var allTemplates = map[string]*DraftConfig{}

var validTemplateTypes = map[string]bool{
	"manifest":   true,
	"dockerfile": true,
	"workflow":   true,
	"deployment": true,
}

var validVariableTypes = map[string]bool{
	"string": true,
	"bool":   true,
	"int":    true,
	"float":  true,
	"object": true,
}
var validVariableKinds = map[string]bool{
	"azureContainerRegistry": true,
	"azureKeyvaultUri":       true,
	"azureManagedCluster":    true,
	"azureResourceGroup":     true,
	"azureServiceConnection": true,
	"containerImageName":     true,
	"containerImageVersion":  true,
	"dirPath":                true,
	"filePath":               true,
	"flag":                   true,
	"helmChartOverrides":     true,
	"ingressHostName":        true,
	"kubernetesNamespace":    true,
	"kubernetesResourceName": true,
	"label":                  true,
	"port":                   true,
	"repositoryBranch":       true,
	"workflowName":           true,
}

/*
This test will validate all the templates in the templates directory have:
1. a unique template name
2. a valid template type
3. a non-empty variable name
4. a valid variable type
5. a valid variable kind

Append this for more validation
*/
func TestTempalteValidation(t *testing.T) {
	assert.Nil(t, loadTemplatesWithValidation())
}

func loadTemplatesWithValidation() error {
	return fs.WalkDir(template.Templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.EqualFold(d.Name(), draftConfigFile) {
			return nil
		}

		currTemplate, err := NewConfigFromFS(template.Templates, path)
		if err != nil {
			return err
		}

		if currTemplate == nil {
			return fmt.Errorf("template %s is nil", path)
		}

		if currTemplate.TemplateName == "" {
			return fmt.Errorf("template %s has no template name", path)
		}

		if _, ok := allTemplates[currTemplate.TemplateName]; ok {
			return fmt.Errorf("template %s has a duplicate template name", path)
		}

		if _, ok := validTemplateTypes[currTemplate.Type]; !ok {
			return fmt.Errorf("template %s has an invalid type: %s", path, currTemplate.Type)
		}

		for _, variable := range currTemplate.Variables {
			if variable.Name == "" {
				return fmt.Errorf("template %s has a variable with no name", path)
			}

			if _, ok := validVariableTypes[variable.Type]; !ok {
				return fmt.Errorf("template %s has an invalid variable(%s) type: %s", path, variable.Name, variable.Type)
			}

			if _, ok := validVariableKinds[variable.Kind]; !ok {
				return fmt.Errorf("template %s has an invalid variable kind: %s", path, variable.Kind)
			}
		}
		return nil
	})
}
