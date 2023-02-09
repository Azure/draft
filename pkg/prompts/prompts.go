package prompts

import (
	"fmt"
	"io"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
)

func RunPromptsFromConfig(config *config.DraftConfig) (map[string]string, error) {
	return RunPromptsFromConfigWithSkips(config, []string{})
}

func RunPromptsFromConfigWithSkips(config *config.DraftConfig, varsToSkip []string) (map[string]string, error) {
	return RunPromptsFromConfigWithSkipsIO(config, varsToSkip, nil, nil)
}

// RunPromptsFromConfigWithSkipsIO runs the prompts for the given config
// skipping any variables in varsToSkip or where the BuilderVar.IsPromptDisabled is true.
// If Stdin or Stdout are nil, the default values will be used.
func RunPromptsFromConfigWithSkipsIO(config *config.DraftConfig, varsToSkip []string, Stdin io.ReadCloser, Stdout io.WriteCloser) (map[string]string, error) {
	skipMap := make(map[string]interface{})
	for _, v := range varsToSkip {
		skipMap[v] = interface{}(nil)
	}

	inputs := make(map[string]string)

	for _, customPrompt := range config.Variables {
		promptVariableName := customPrompt.Name
		if _, ok := skipMap[promptVariableName]; ok {
			log.Debugf("Skipping prompt for %s", promptVariableName)
			continue
		}
		if customPrompt.IsPromptDisabled {
			log.Debugf("Skipping prompt for %s as it has IsPromptDisabled=true", promptVariableName)
			noPromptDefaultValue := GetVariableDefaultValue(promptVariableName, config.VariableDefaults, inputs)
			if noPromptDefaultValue == "" {
				return nil, fmt.Errorf("IsPromptDisabled is true for %s but no default value was found", promptVariableName)
			}
			log.Debugf("Using default value %s for %s", noPromptDefaultValue, promptVariableName)
			inputs[promptVariableName] = noPromptDefaultValue
			continue
		}

		log.Debugf("constructing prompt for: %s", promptVariableName)
		if customPrompt.VarType == "bool" {
			input, err := RunBoolPrompt(customPrompt, Stdin, Stdout)
			if err != nil {
				return nil, err
			}
			inputs[promptVariableName] = input
		} else {
			defaultValue := GetVariableDefaultValue(promptVariableName, config.VariableDefaults, inputs)

			stringInput, err := RunDefaultableStringPrompt(customPrompt, defaultValue, nil, Stdin, Stdout)
			if err != nil {
				return nil, err
			}
			inputs[promptVariableName] = stringInput
		}
	}

	// Substitute the default value for variables where the user didn't enter anything
	for _, variableDefault := range config.VariableDefaults {
		if inputs[variableDefault.Name] == "" {
			inputs[variableDefault.Name] = variableDefault.Value
		}
	}

	return inputs, nil
}

// GetVariableDefaultValue returns the default value for a variable, if one is set in variableDefaults from a ReferenceVar or literal VariableDefault.Value in that order.
func GetVariableDefaultValue(variableName string, variableDefaults []config.BuilderVarDefault, inputs map[string]string) string {
	defaultValue := ""
	for _, variableDefault := range variableDefaults {
		if variableDefault.Name == variableName {
			defaultValue = variableDefault.Value
			log.Debugf("setting default value for %s to %s from variable default rule", variableName, defaultValue)
			if variableDefault.ReferenceVar != "" && inputs[variableDefault.ReferenceVar] != "" {
				defaultValue = inputs[variableDefault.ReferenceVar]
				log.Debugf("setting default value for %s to %s from referenceVar %s", variableName, defaultValue, variableDefault.ReferenceVar)
			}
		}
	}
	return defaultValue
}

func RunBoolPrompt(customPrompt config.BuilderVar, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
	newSelect := &promptui.Select{
		Label:  "Please select " + customPrompt.Description,
		Items:  []bool{true, false},
		Stdin:  Stdin,
		Stdout: Stdout,
	}

	_, input, err := newSelect.Run()
	if err != nil {
		return "", err
	}
	return input, nil
}

// AllowAllStringValidator is a string validator that allows any string
func AllowAllStringValidator(_ string) error {
	return nil
}

// NoBlankStringValidator is a string validator that does not allow blank strings
func NoBlankStringValidator(s string) error {
	if len(s) <= 0 {
		return fmt.Errorf("input must be greater than 0")
	}
	return nil
}

// RunDefaultableStringPrompt runs a prompt for a string variable, returning the user string input for the prompt
func RunDefaultableStringPrompt(customPrompt config.BuilderVar, defaultValue string, validate func(string) error, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
	var validatorFunc func(string) error
	if validate == nil {
		validatorFunc = NoBlankStringValidator
	}

	defaultString := ""
	if defaultValue != "" {
		validatorFunc = AllowAllStringValidator
		defaultString = " (default: " + defaultValue + ")"
	}

	prompt := &promptui.Prompt{
		Label:    "Please enter " + customPrompt.Description + defaultString,
		Validate: validatorFunc,
		Stdin:    Stdin,
		Stdout:   Stdout,
	}

	input, err := prompt.Run()
	if err != nil {
		return "", err
	}
	// Variable-level substitution, we need to get defaults so later references can be resolved in this loop
	if input == "" && defaultString != "" {
		input = defaultValue
	}
	return input, nil
}

func GetInputFromPrompt(desiredInput string) string {
	prompt := &promptui.Prompt{
		Label: "Please enter " + desiredInput,
		Validate: func(s string) error {
			if len(s) <= 0 {
				return fmt.Errorf("input must be greater than 0")
			}
			return nil
		},
	}

	input, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}

	return input
}
