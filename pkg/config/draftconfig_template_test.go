package config

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/draft/template"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

const alphaNumUnderscoreHyphen = "^[A-Za-z][A-Za-z0-9-_]{1,62}[A-Za-z0-9]$"

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
	"azureContainerRegistry":     true,
	"azureKeyvaultUri":           true,
	"azureManagedCluster":        true,
	"azureResourceGroup":         true,
	"azureServiceConnection":     true,
	"containerImageName":         true,
	"containerImageVersion":      true,
	"dirPath":                    true,
	"dockerFileName":             true,
	"filePath":                   true,
	"flag":                       true,
	"helmChartOverrides":         true,
	"ingressHostName":            true,
	"kubernetesNamespace":        true,
	"kubernetesResourceName":     true,
	"label":                      true,
	"port":                       true,
	"repositoryBranch":           true,
	"workflowName":               true,
	"replicaCount":               true,
	"scalingResourceType":        true,
	"scalingResourceUtilization": true,
	"resourceLimit":              true,
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
	regexp := regexp.MustCompile(alphaNumUnderscoreHyphen)
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

		if !regexp.MatchString(currTemplate.TemplateName) {
			return fmt.Errorf("template %s name must match the alpha-numeric-underscore-hyphen regex: %s", path, currTemplate.TemplateName)
		}

		if _, ok := allTemplates[strings.ToLower(currTemplate.TemplateName)]; ok {
			return fmt.Errorf("template %s has a duplicate template name", path)
		}

		if _, ok := validTemplateTypes[currTemplate.Type]; !ok {
			return fmt.Errorf("template %s has an invalid type: %s", path, currTemplate.Type)
		}

		if _, err := semver.ParseRange(currTemplate.Versions); err != nil {
			return fmt.Errorf("template %s has an invalid version range: %s", path, currTemplate.Versions)
		}

		referenceVarMap := map[string]*BuilderVar{}
		conditionRefMap := map[string]*BuilderVar{}
		allVariables := map[string]*BuilderVar{}
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

			if _, err := semver.ParseRange(variable.Versions); err != nil {
				return fmt.Errorf("template %s has an invalid version range: %s", path, variable.Versions)
			}

			allVariables[variable.Name] = variable
			if variable.Default.ReferenceVar != "" {
				referenceVarMap[variable.Name] = variable
			}

			if variable.ConditionalRef.ReferenceVar != "" {
				conditionRefMap[variable.Name] = variable
			}
		}

		for _, currVar := range referenceVarMap {
			refVar, ok := allVariables[currVar.Default.ReferenceVar]
			if !ok {
				return fmt.Errorf("template %s has a variable %s with default reference to a non-existent variable: %s", path, currVar.Name, currVar.Default.ReferenceVar)
			}

			if currVar.Name == refVar.Name {
				return fmt.Errorf("template %s has a variable with cyclical default reference to itself: %s", path, currVar.Name)
			}

			if isCyclicalDefaultVariableReference(currVar, refVar, allVariables, map[string]bool{}) {
				return fmt.Errorf("template %s has a variable with cyclical default reference to itself: %s", path, currVar.Name)
			}
		}

		for _, currVar := range conditionRefMap {
			refVar, ok := allVariables[currVar.ConditionalRef.ReferenceVar]
			if !ok {
				return fmt.Errorf("template %s has a variable %s with conditional reference to a non-existent variable: %s", path, currVar.Name, currVar.ConditionalRef.ReferenceVar)
			}

			if isCyclicalConditionalVariableReference(currVar, refVar, allVariables, map[string]bool{}) {
				return fmt.Errorf("template %s has a variable with cyclical conditional reference to itself or references a non existing variable: %s", path, currVar.Name)
			}
		}

		allTemplates[strings.ToLower(currTemplate.TemplateName)] = currTemplate
		return nil
	})
}

func isCyclicalDefaultVariableReference(initialVar, currRefVar *BuilderVar, allVariables map[string]*BuilderVar, visited map[string]bool) bool {
	if initialVar.Name == currRefVar.Name {
		return true
	}

	if _, ok := visited[currRefVar.Name]; ok {
		return true
	}

	if currRefVar.Default.ReferenceVar == "" {
		return false
	}

	refVar, ok := allVariables[currRefVar.Default.ReferenceVar]
	if !ok {
		return false
	}

	visited[currRefVar.Name] = true
	return isCyclicalDefaultVariableReference(initialVar, refVar, allVariables, visited)
}

func isCyclicalConditionalVariableReference(initialVar, currRefVar *BuilderVar, allVariables map[string]*BuilderVar, visited map[string]bool) bool {
	if initialVar.Name == currRefVar.Name {
		return true
	}

	if _, ok := visited[currRefVar.Name]; ok {
		return true
	}

	if currRefVar.ConditionalRef.ReferenceVar == "" {
		return false
	}

	refVar, ok := allVariables[currRefVar.ConditionalRef.ReferenceVar]
	if !ok {
		return false
	}

	visited[currRefVar.Name] = true
	return isCyclicalConditionalVariableReference(initialVar, refVar, allVariables, visited)
}
