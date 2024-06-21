package config

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// TODO: remove Name Overrides since we don't need them anymore
type DraftConfig struct {
	DisplayName   string             `yaml:"displayName"`
	NameOverrides []FileNameOverride `yaml:"nameOverrides"`
	Variables     []BuilderVar       `yaml:"variables"`

	nameOverrideMap map[string]string
}

type FileNameOverride struct {
	Path   string `yaml:"path"`
	Prefix string `yaml:"prefix"`
}

type BuilderVar struct {
	Name          string            `yaml:"name"`
	Default       BuilderVarDefault `yaml:"default"`
	Description   string            `yaml:"description"`
	ExampleValues []string          `yaml:"exampleValues"`
	Type          string            `yaml:"type"`
}

type BuilderVarDefault struct {
	IsPromptDisabled bool   `yaml:"disablePrompt"`
	ReferenceVar     string `yaml:"referenceVar"`
	Value            string `yaml:"value"`
}

func (d *DraftConfig) GetVariableExampleValues() map[string][]string {
	variableExampleValues := make(map[string][]string)
	for _, variable := range d.Variables {
		if len(variable.ExampleValues) > 0 {
			variableExampleValues[variable.Name] = variable.ExampleValues
		}
	}

	return variableExampleValues
}

func (d *DraftConfig) initNameOverrideMap() {
	d.nameOverrideMap = make(map[string]string)
	log.Debug("initializing nameOverrideMap")
	for _, builderVar := range d.NameOverrides {
		log.Debugf("mapping path: %s, to prefix %s", builderVar.Path, builderVar.Prefix)
		d.nameOverrideMap[builderVar.Path] = builderVar.Prefix
	}
}

func (d *DraftConfig) GetNameOverride(path string) string {
	if d.nameOverrideMap == nil {
		d.initNameOverrideMap()
	}
	prefix, ok := d.nameOverrideMap[path]
	if !ok {
		return ""
	}

	return prefix
}

// ApplyDefaultVariables will apply the defaults to variables that are not already set
func (d *DraftConfig) ApplyDefaultVariables(customInputs map[string]string) error {
	for _, variable := range d.Variables {
		// handle where variable is not set or is set to an empty string from cli handling
		if customInputs[variable.Name] == "" {
			if variable.Default.ReferenceVar != "" {
				customInputs[variable.Name] = d.RecurseReferenceVars(variable, customInputs, make(map[string]int))
				log.Infof("Variable %s defaulting to value %s", variable.Name, customInputs[variable.Name])
			}

			if customInputs[variable.Name] == "" {
				if variable.Default.Value != "" {
					log.Infof("Variable %s defaulting to value %s", variable.Name, variable.Default.Value)
					customInputs[variable.Name] = variable.Default.Value
				} else {
					return errors.New("variable " + variable.Name + " has no default value")
				}
			}
		}
	}

	return nil
}

func (d *DraftConfig) RecurseReferenceVars(variable BuilderVar, customInputs map[string]string, varIdxMap map[string]int) string {
	if customInputs[variable.Default.ReferenceVar] != "" {
		return customInputs[variable.Default.ReferenceVar]
	} else if variable.Default.ReferenceVar != "" {
		return d.RecurseReferenceVars(d.Variables[varIdxMap[variable.Default.ReferenceVar]], customInputs, varIdxMap)
	}

	return variable.Default.Value
}

// TemplateVariableRecorder is an interface for recording variables that are used read using draft configs
type TemplateVariableRecorder interface {
	Record(key, value string)
}
