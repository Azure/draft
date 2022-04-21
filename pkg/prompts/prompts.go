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

func RunPromptsFromConfig(config *config.DraftConfig) (map[string]string, error) {
	templatePrompts := make([]*TemplatePrompt, 0)
	for _, customPrompt := range config.Variables {
		prompt := &promptui.Prompt{
			Label: "Please Enter " + customPrompt.Description,
			Validate: func(s string) error {
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

	inputs := make(map[string]string)

	for _, prompt := range templatePrompts {
		input, err := prompt.Prompt.Run()
		if err != nil {
			return nil, err
		}
		inputs[prompt.OverrideString] = input
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
