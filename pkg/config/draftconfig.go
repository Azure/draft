package config

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// TODO: remove Name Overrides since we don't need them anymore
type DraftConfig struct {
	DisplayName   string                `yaml:"displayName"`
	NameOverrides []FileNameOverride    `yaml:"nameOverrides"`
	Variables     map[string]BuilderVar `yaml:"variables"`

	nameOverrideMap map[string]string
}

type FileNameOverride struct {
	Path   string `yaml:"path"`
	Prefix string `yaml:"prefix"`
}

type BuilderVar struct {
	DefaultValue     string   `yaml:"default"`
	Description      string   `yaml:"description"`
	ExampleValues    []string `yaml:"exampleValues"`
	IsPromptDisabled bool     `yaml:"disablePrompt"`
	ReferenceVar     string   `yaml:"referenceVar"`
	Type             string   `yaml:"type"`
	Value            string   `yaml:"value"`
}

func (d *DraftConfig) GetVariableExampleValues() map[string][]string {
	variableExampleValues := make(map[string][]string)
	for name, variable := range d.Variables {
		if len(variable.ExampleValues) > 0 {
			variableExampleValues[name] = variable.ExampleValues
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
func (d *DraftConfig) ApplyDefaultVariables(customConfig map[string]string) error {
	for name, variable := range d.Variables {
		// handle where variable is not set or is set to an empty string from cli handling
		if val, ok := customConfig[name]; !ok || val == "" {
			if variable.DefaultValue == "" && name != "DOCKERFILE" {
				return errors.New("variable " + name + " has no default value")
			}
			log.Infof("Variable %s defaulting to value %s", name, variable.DefaultValue)
			customConfig[name] = variable.DefaultValue
		}
	}

	return nil
}

// TemplateVariableRecorder is an interface for recording variables that are used read using draft configs
type TemplateVariableRecorder interface {
	Record(key, value string)
}

func (d *DraftConfig) VariableMap() (map[string]string, error) {
	varMap := make(map[string]string)
	for name, variable := range d.Variables {
		if variable.Value == "" {
			return nil, errors.New("variable " + name + " has no default value")
		}
		varMap[name] = variable.Value
	}

	return varMap, nil
}
