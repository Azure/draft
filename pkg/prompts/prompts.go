package prompts

import (
	"fmt"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
)

type TemplatePrompt struct {
	Prompt         *promptui.Prompt
	OverrideString string
}

type TemplateSelect struct {
	Select         *promptui.Select
	OverrideString string
}

func RunPromptsFromConfig(config *config.DraftConfig) (map[string]string, error) {
	return RunPromptsFromConfigWithSkips(config, []string{})
}

func RunPromptsFromConfigWithSkips(config *config.DraftConfig, varsToSkip []string) (map[string]string, error) {
	skipMap := make(map[string]interface{})
	for _, v := range varsToSkip {
		skipMap[v] = interface{}(nil)
	}

	inputs := make(map[string]string)

	for _, customPrompt := range config.Variables {
		if _, ok := skipMap[customPrompt.Name]; ok {
			log.Debugf("Skipping prompt for %s", customPrompt.Name)
			continue
		}

		log.Debugf("constructing prompt for: %s", customPrompt)
		if customPrompt.VarType == "bool" {
			prompt := &promptui.Select{
				Label: "Please select " + customPrompt.Description,
				Items: []bool{true, false},
			}

			// TODO: we probably don't need this struct anymore since we are getting rid of override strings
			newSelect := &TemplateSelect{
				Select:         prompt,
				OverrideString: customPrompt.Name,
			}

			_, input, err := newSelect.Select.Run()
			if err != nil {
				return nil, err
			}
			inputs[newSelect.OverrideString] = input
		} else {
			defaultString := ""
			defaultValue := ""
			for _, variableDefault := range config.VariableDefaults {
				if variableDefault.Name == customPrompt.Name {
					defaultValue = variableDefault.Value
					if variableDefault.ReferenceVar != "" {
						defaultValue = inputs[variableDefault.ReferenceVar]
						log.Debugf("setting default value for %s to %s from referenceVar %s", customPrompt.Name, defaultValue, variableDefault.ReferenceVar)
					}
					defaultString = " (default: " + defaultValue + ")"
				}
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
			}

			// TODO: we probably don't need this struct anymore since we are getting rid of override strings
			newPrompt := &TemplatePrompt{
				Prompt:         prompt,
				OverrideString: customPrompt.Name,
			}
			input, err := newPrompt.Prompt.Run()
			if err != nil {
				return nil, err
			}
			// Variable-level substitution, we need to get defaults so later references can be resolved in this loop
			if input == "" && defaultString != "" {
				input = defaultValue
			}
			inputs[newPrompt.OverrideString] = input
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
