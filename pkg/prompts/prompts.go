package prompts

import "github.com/manifoldco/promptui"

type TemplatePrompt struct {
	Prompt *promptui.Prompt
	OverrideString string
}