package config

import (
	log "github.com/sirupsen/logrus"
)

// TODO: remove Name Overrides since we don't need them anymore
type DraftConfig struct {
	DisplayName      string              `yaml:"displayName"`
	NameOverrides    []FileNameOverride  `yaml:"nameOverrides"`
	Variables        []BuilderVar        `yaml:"variables"`
	VariableDefaults []BuilderVarDefault `yaml:"variableDefaults"`

	nameOverrideMap map[string]string
}

type FileNameOverride struct {
	Path   string `yaml:"path"`
	Prefix string `yaml:"prefix"`
}

type BuilderVar struct {
	Name             string   `yaml:"name"`
	Description      string   `yaml:"description"`
	VarType          string   `yaml:"type"`
	ExampleValues    []string `yaml:"exampleValues"`
}

type BuilderVarDefault struct {
	Name             string `yaml:"name"`
	Value            string `yaml:"value"`
	ReferenceVar     string `yaml:"referenceVar"`
	IsPromptDisabled bool   `yaml:"disablePrompt"`
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
func (d *DraftConfig) ApplyDefaultVariables(customConfig map[string]string) {
	for _, variable := range d.VariableDefaults {
		// handle where variable is not set or is set to an empty string from cli handling
		if defaultVal, ok := customConfig[variable.Name]; !ok || defaultVal == "" {
			log.Infof("Variable %s defaulting to value %s", variable.Name, variable.Value)
			customConfig[variable.Name] = variable.Value
		}
	}
}

// TemplateVariableRecorder is an interface for recording variables that are used read using draft configs
type TemplateVariableRecorder interface {
	Record(key, value string)
}
