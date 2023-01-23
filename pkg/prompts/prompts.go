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

func RunPromptsFromConfigWithSkipsIO(config *config.DraftConfig, varsToSkip []string, Stdin io.ReadCloser, Stdout io.WriteCloser) (map[string]string, error) {
	skipMap := make(map[string]interface{})
	for _, v := range varsToSkip {
		skipMap[v] = interface{}(nil)
	}

	inputs := make(map[string]string)

	for _, customPrompt := range config.Variables {
		promptVariableName := customPrompt.Name
		if _, ok := skipMap[customPrompt.Name]; ok {
			log.Debugf("Skipping prompt for %s", customPrompt.Name)
			continue
		}

		log.Debugf("constructing prompt for: %s", customPrompt)
		if customPrompt.VarType == "bool" {
			input, err := RunBoolPrompt(customPrompt, Stdin, Stdout)
			if err != nil {
				return nil, err
			}
			inputs[promptVariableName] = input
		} else {
			defaultValue := GetVariableDefaultValue(promptVariableName, config.VariableDefaults, inputs)

			stringInput, err := RunStringPrompt(customPrompt, defaultValue, Stdin, Stdout)
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
			if variableDefault.ReferenceVar != "" {
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

// RunStringPrompt runs a prompt for a string variable, returning the user string input for the prompt
func RunStringPrompt(customPrompt config.BuilderVar, defaultValue string, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
	defaultString := ""
	if defaultValue != "" {
		defaultString = " (default: " + defaultValue + ")"
	}

	prompt := &promptui.Prompt{
		Label: "Please enter " + customPrompt.Description + defaultString,
		Validate: func(s string) error {
			// Allow blank input for variables with defaults
			if defaultString != "" {
				return nil
			}
			if len(s) <= 0 {
				return fmt.Errorf("input must be greater than 0")
			}
			return nil
		},
		Stdin:  Stdin,
		Stdout: Stdout,
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
