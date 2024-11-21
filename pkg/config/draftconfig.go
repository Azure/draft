package config

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/Azure/draft/pkg/config/transformers"
	"github.com/Azure/draft/pkg/config/validators"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/blang/semver/v4"
)

type VariableCondition string

const (
	EqualTo    VariableCondition = "equals"
	NotEqualTo VariableCondition = "notequals"
)

func (v VariableCondition) String() string {
	return string(v)
}

const draftConfigFile = "draft.yaml"

type VariableValidator func(string) error
type VariableTransformer func(string) (any, error)

type DraftConfig struct {
	TemplateName        string                         `yaml:"templateName"`
	DisplayName         string                         `yaml:"displayName"`
	Description         string                         `yaml:"description"`
	Type                string                         `yaml:"type"`
	Versions            string                         `yaml:"versions"`
	DefaultVersion      string                         `yaml:"defaultVersion"`
	Variables           []*BuilderVar                  `yaml:"variables"`
	FileNameOverrideMap map[string]string              `yaml:"filenameOverrideMap"`
	Validators          map[string]VariableValidator   `yaml:"validators"`
	Transformers        map[string]VariableTransformer `yaml:"transformers"`
}

type BuilderVar struct {
	Name                  string                 `yaml:"name"`
	ActiveWhenConstraints []ActiveWhenConstraint `yaml:"activeWhen"`
	Default               BuilderVarDefault      `yaml:"default"`
	Description           string                 `yaml:"description"`
	ExampleValues         []string               `yaml:"exampleValues"`
	AllowedValues         []string               `yaml:"allowedValues"`
	Type                  string                 `yaml:"type"`
	Kind                  string                 `yaml:"kind"`
	Value                 string                 `yaml:"value"`
	Versions              string                 `yaml:"versions"`
}

// BuilderVarDefault holds info on the default value of a variable
type BuilderVarDefault struct {
	IsPromptDisabled bool   `yaml:"disablePrompt"`
	ReferenceVar     string `yaml:"referenceVar"`
	Value            string `yaml:"value"`
}

// ActiveWhenConstraints holds information on when a variable is actively used by a template based off other variable values
type ActiveWhenConstraint struct {
	VariableName string            `yaml:"variableName"`
	Value        string            `yaml:"value"`
	Condition    VariableCondition `yaml:"condition"`
}

func NewConfigFromFS(fileSys fs.FS, path string) (*DraftConfig, error) {
	configBytes, err := fs.ReadFile(fileSys, path)
	if err != nil {
		return nil, err
	}

	var draftConfig DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
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

// Returns a map of variable names to values used in Gotemplate
func (d *DraftConfig) GetVariableMap() map[string]string {
	variableMap := make(map[string]string)
	for _, variable := range d.Variables {
		variableMap[variable.Name] = variable.Value
	}
	return variableMap
}

func (d *DraftConfig) GetVariable(name string) (*BuilderVar, error) {
	for _, variable := range d.Variables {
		if variable.Name == name {
			return variable, nil
		}
	}

	return nil, fmt.Errorf("variable %s not found", name)
}

func (d *DraftConfig) GetVariableValue(name string) (any, error) {
	for _, variable := range d.Variables {
		if variable.Name == name {
			if variable.Value == "" {
				return "", fmt.Errorf("variable %s has no value", name)
			}

			if err := d.GetVariableValidator(variable.Kind)(variable.Value); err != nil {
				return "", fmt.Errorf("failed variable validation: %w", err)
			}

			response, err := d.GetVariableTransformer(variable.Kind)(variable.Value)
			if err != nil {
				return "", fmt.Errorf("failed variable transformation: %w", err)
			}

			return response, nil
		}
	}

	return "", fmt.Errorf("variable %s not found", name)
}

func (d *DraftConfig) SetVariable(name, value string) {
	if variable, err := d.GetVariable(name); err != nil {
		d.Variables = append(d.Variables, &BuilderVar{
			Name:  name,
			Value: value,
		})
	} else {
		variable.Value = value
	}
}

// GetVariableTransformer returns the transformer for a specific variable kind
func (d *DraftConfig) GetVariableTransformer(kind string) VariableTransformer {
	// user overrides
	if transformer, ok := d.Transformers[kind]; ok {
		return transformer
	}

	// internally defined transformers
	return transformers.GetTransformer(kind)
}

// GetVariableValidator returns the validator for a specific variable kind
func (d *DraftConfig) GetVariableValidator(kind string) VariableValidator {
	// user overrides
	if validator, ok := d.Validators[kind]; ok {
		return validator
	}

	// internally defined validators
	return validators.GetValidator(kind)
}

// SetVariableTransformer sets the transformer for a specific variable kind
func (d *DraftConfig) SetVariableTransformer(kind string, transformer VariableTransformer) {
	if d.Transformers == nil {
		d.Transformers = make(map[string]VariableTransformer)
	}
	d.Transformers[kind] = transformer
}

// SetVariableValidator sets the validator for a specific variable kind
func (d *DraftConfig) SetVariableValidator(kind string, validator VariableValidator) {
	if d.Validators == nil {
		d.Validators = make(map[string]VariableValidator)
	}
	d.Validators[kind] = validator
}

// ApplyDefaultVariables will apply the defaults to variables that are not already set
func (d *DraftConfig) ApplyDefaultVariables() error {
	for _, variable := range d.Variables {
		if variable.Value == "" {
			if variable.Default.ReferenceVar != "" {
				referenceVar, err := d.GetVariable(variable.Default.ReferenceVar)
				if err != nil {
					return fmt.Errorf("apply default variables: %w", err)
				}
				defaultVal, err := d.recurseReferenceVars(referenceVar, referenceVar, true)
				if err != nil {
					return fmt.Errorf("apply default variables: %w", err)
				}
				log.Infof("Variable %s defaulting to value %s", variable.Name, defaultVal)
				variable.Value = defaultVal
			}

			isVarActive, err := d.CheckActiveWhenConstraint(variable)
			if err != nil {
				return fmt.Errorf("unable to check ActiveWhen constraint: %w", err)
			}

			if !isVarActive {
				continue
			}

			if variable.Value == "" {
				if variable.Default.Value != "" {
					log.Infof("Variable %s defaulting to value %s", variable.Name, variable.Default.Value)
					variable.Value = variable.Default.Value
				} else {
					return errors.New("variable " + variable.Name + " has no default value")
				}
			}
		}
	}

	return nil
}

// ApplyDefaultVariablesForVersion will apply the defaults to variables that are not already set for a specific template version
func (d *DraftConfig) ApplyDefaultVariablesForVersion(version string) error {
	v, err := semver.Parse(version)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	expectedConfigVersionRange, err := semver.ParseRange(d.Versions)
	if err != nil {
		return fmt.Errorf("invalid config version range: %w", err)
	}

	if !expectedConfigVersionRange(v) {
		return fmt.Errorf("version %s is outside of config version range %s", version, d.Versions)
	}

	for _, variable := range d.Variables {
		if variable.Value == "" {
			expectedRange, err := semver.ParseRange(variable.Versions)
			if err != nil {
				return fmt.Errorf("invalid variable versions: %w", err)
			}

			if !expectedRange(v) {
				log.Infof("Variable %s versions %s is outside input version %s, skipping", variable.Name, variable.Versions, version)
				continue
			}

			if variable.Default.ReferenceVar != "" {
				referenceVar, err := d.GetVariable(variable.Default.ReferenceVar)
				if err != nil {
					return fmt.Errorf("apply default variables: %w", err)
				}

				defaultVal, err := d.recurseReferenceVars(referenceVar, referenceVar, true)
				if err != nil {
					return fmt.Errorf("apply default variables: %w", err)
				}
				log.Infof("Variable %s defaulting to value %s", variable.Name, defaultVal)
				variable.Value = defaultVal
			}

			isVarActive, err := d.CheckActiveWhenConstraint(variable)
			if err != nil {
				return fmt.Errorf("unable to check ActiveWhen constraint: %w", err)
			}

			if !isVarActive {
				continue
			}

			if variable.Value == "" {
				if variable.Default.Value != "" {
					log.Infof("Variable %s defaulting to value %s", variable.Name, variable.Default.Value)
					variable.Value = variable.Default.Value
				} else {
					return errors.New("variable " + variable.Name + " has no default value")
				}
			}
		}
	}

	return nil
}

func (d *DraftConfig) CheckActiveWhenConstraint(variable *BuilderVar) (bool, error) {
	if len(variable.ActiveWhenConstraints) > 0 {
		isVarActive := true
		for _, activeWhen := range variable.ActiveWhenConstraints {
			refVar, err := d.GetVariable(activeWhen.VariableName)
			if err != nil {
				return false, fmt.Errorf("unable to get ActiveWhen reference variable: %w", err)
			}

			checkValue := refVar.Value
			if checkValue == "" {
				if refVar.Default.Value != "" {
					checkValue = refVar.Default.Value
				}

				if refVar.Default.ReferenceVar != "" {
					refValue, err := d.recurseReferenceVars(refVar, refVar, true)
					if err != nil {
						return false, err
					}
					if refValue == "" {
						return false, errors.New("reference variable has no value")
					}

					checkValue = refValue
				}
			}

			switch activeWhen.Condition {
			case EqualTo:
				isVarActive = checkValue == activeWhen.Value
			case NotEqualTo:
				isVarActive = checkValue != activeWhen.Value
			default:
				return false, fmt.Errorf("invalid activeWhen condition: %s", activeWhen.Condition)
			}
		}
		return isVarActive, nil
	}

	return true, nil
}

// recurseReferenceVars recursively checks each variable's ReferenceVar if it doesn't have a custom input. If there's no more ReferenceVars, it will return the default value of the last ReferenceVar.
func (d *DraftConfig) recurseReferenceVars(referenceVar *BuilderVar, variableCheck *BuilderVar, isFirst bool) (string, error) {
	if !isFirst && referenceVar.Name == variableCheck.Name {
		return "", errors.New("cyclical reference detected")
	}

	// If referenceVar has a custom value, return it, else check its ReferenceVar, else return its default value
	if referenceVar.Value != "" {
		return referenceVar.Value, nil
	} else if referenceVar.Default.ReferenceVar != "" {
		referenceVar, err := d.GetVariable(referenceVar.Default.ReferenceVar)
		if err != nil {
			return "", fmt.Errorf("recurse reference vars: %w", err)
		}

		return d.recurseReferenceVars(referenceVar, variableCheck, false)
	}

	return referenceVar.Default.Value, nil
}

// handles flags that are meant to represent template variables
func (d *DraftConfig) VariableMapToDraftConfig(flagVariablesMap map[string]string) {
	for flagName, flagValue := range flagVariablesMap {
		log.Debugf("flag variable %s=%s", flagName, flagValue)
		d.SetVariable(flagName, flagValue)
	}
}

// SetFileNameOverride sets the filename override for a specific file
func (d *DraftConfig) SetFileNameOverride(input, override string) {
	if d.FileNameOverrideMap == nil {
		d.FileNameOverrideMap = make(map[string]string)
	}
	d.FileNameOverrideMap[input] = override
}

func (d *DraftConfig) DeepCopy() *DraftConfig {
	newConfig := &DraftConfig{
		TemplateName:        d.TemplateName,
		DisplayName:         d.DisplayName,
		Description:         d.Description,
		Type:                d.Type,
		Versions:            d.Versions,
		DefaultVersion:      d.DefaultVersion,
		Variables:           make([]*BuilderVar, len(d.Variables)),
		FileNameOverrideMap: make(map[string]string),
	}
	for i, variable := range d.Variables {
		newConfig.Variables[i] = variable.DeepCopy()
	}

	for k, v := range d.FileNameOverrideMap {
		newConfig.FileNameOverrideMap[k] = v
	}

	return newConfig
}

func (bv *BuilderVar) DeepCopy() *BuilderVar {
	newVar := &BuilderVar{
		Name:                  bv.Name,
		Default:               bv.Default,
		Description:           bv.Description,
		Type:                  bv.Type,
		Kind:                  bv.Kind,
		Value:                 bv.Value,
		Versions:              bv.Versions,
		ExampleValues:         make([]string, len(bv.ExampleValues)),
		AllowedValues:         make([]string, len(bv.AllowedValues)),
		ActiveWhenConstraints: make([]ActiveWhenConstraint, len(bv.ActiveWhenConstraints)),
	}
	for i, awc := range bv.ActiveWhenConstraints {
		newVar.ActiveWhenConstraints[i] = *awc.DeepCopy()
	}
	copy(newVar.AllowedValues, bv.AllowedValues)
	copy(newVar.ExampleValues, bv.ExampleValues)
	return newVar
}

func (awc ActiveWhenConstraint) DeepCopy() *ActiveWhenConstraint {
	return &ActiveWhenConstraint{
		VariableName: awc.VariableName,
		Value:        awc.Value,
		Condition:    awc.Condition,
	}
}

// TemplateVariableRecorder is an interface for recording variables that are read using draft configs
type TemplateVariableRecorder interface {
	Record(key, value string)
}
