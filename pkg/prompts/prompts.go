package prompts

import (
	"fmt"

	"github.com/Azure/draft/pkg/config"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
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
	templatePrompts := make([]*TemplatePrompt, 0)
	templateSelects := make([]*TemplateSelect, 0)

	skipMap := make(map[string]interface{})
	for _, v := range varsToSkip {
		skipMap[v] = interface{}(nil)
	}

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

			templateSelects = append(templateSelects, &TemplateSelect{
				Select:         prompt,
				OverrideString: customPrompt.Name,
			})
		} else {
			defaultString := ""
			for _, variableDefault := range config.VariableDefaults {
				if variableDefault.Name == customPrompt.Name {
					defaultString = " (default: " + variableDefault.Value + ")"
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

			templatePrompts = append(templatePrompts, &TemplatePrompt{
				Prompt:         prompt,
				OverrideString: customPrompt.Name,
			})
		}
	}

	inputs := make(map[string]string)

	for _, prompt := range templatePrompts {
		input, err := prompt.Prompt.Run()
		if err != nil {
			return nil, err
		}
		inputs[prompt.OverrideString] = input
	}

	for _, prompt := range templateSelects {
		_, input, err := prompt.Select.Run()
		if err != nil {
			return nil, err
		}
		inputs[prompt.OverrideString] = input
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
