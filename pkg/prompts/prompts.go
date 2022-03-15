package prompts

import (
	"fmt"

	"github.com/Azure/draftv2/pkg/configs"
	"github.com/manifoldco/promptui"
)

type TemplatePrompt struct {
	Prompt         *promptui.Prompt
	OverrideString string
}

func RunPromptsFromConfig(config *configs.DraftConfig) (map[string]string, error) {
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
