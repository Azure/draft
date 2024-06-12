package prompts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
)

const defaultAppName = "my-app"

// Function to get current directory name
var getCurrentDirNameFunc = getCurrentDirName

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
		if GetIsPromptDisabled(customPrompt.Name, config.VariableDefaults) {
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

	if variableName == "APPNAME" {
		dirName, err := getCurrentDirNameFunc()
		if err != nil {
			log.Errorf("Error retrieving current directory name: %s", err)
			return defaultAppName
		}
		defaultValue = sanitizeAppName(dirName)
		return defaultValue
	}

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

func GetIsPromptDisabled(variableName string, variableDefaults []config.BuilderVarDefault) bool {
	for _, variableDefault := range variableDefaults {
		if variableDefault.Name == variableName {
			return variableDefault.IsPromptDisabled
		}
	}
	return false
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

// Validator for App name
func appNameValidator(name string) error {
	if name == "" {
		return fmt.Errorf("application name cannot be empty")
	}

	if !unicode.IsLetter(rune(name[0])) && !unicode.IsDigit(rune(name[0])) {
		return fmt.Errorf("application name must start with a letter or digit")
	}

	if name[len(name)-1] == '-' || name[len(name)-1] == '_' || name[len(name)-1] == '.' {
		return fmt.Errorf("application name must end with a letter or digit")
	}

	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.' {
			return fmt.Errorf("application name can only contain letters, digits, '-', '_', and '.'")
		}
	}

	if len(name) > 63 {
		return fmt.Errorf("application name cannot be longer than 63 characters")
	}

	return nil
}

// RunDefaultableStringPrompt runs a prompt for a string variable, returning the user string input for the prompt
func RunDefaultableStringPrompt(customPrompt config.BuilderVar, defaultValue string, validate func(string) error, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
	if validate == nil {
		validate = NoBlankStringValidator
	}

	validatorFunc := func(input string) error {
		// Allow blank inputs because defaults are set later
		if input == "" {
			return nil
		}
		if customPrompt.Name == "APPNAME" {
			if err := appNameValidator(input); err != nil {
				return err
			}
		} else {
			if err := validate(input); err != nil {
				return err
			}
		}
		return nil
	}

	defaultString := ""
	if defaultValue != "" {
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

	if input == "" && defaultValue != "" {
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

type SelectOpt[T any] struct {
	// Field returns the name to use for each select item.
	Field func(t T) string
	// Default is the default selection. If Field is used this should be the result of calling Field on the default.
	Default *T
}

func Select[T any](label string, items []T, opt *SelectOpt[T]) (T, error) {
	selections := make([]interface{}, len(items))
	for i, item := range items {
		selections[i] = item
	}

	if opt != nil && opt.Field != nil {
		for i, item := range items {
			selections[i] = opt.Field(item)
		}
	}

	if len(selections) == 0 {
		return *new(T), errors.New("no selection options")
	}

	if _, ok := selections[0].(string); !ok {
		return *new(T), errors.New("selections must be of type string or use opt.Field")
	}

	searcher := func(search string, i int) bool {
		str, _ := selections[i].(string) // no need to check if okay, we guard earlier

		selection := strings.ToLower(str)
		search = strings.ToLower(search)

		return strings.Contains(selection, search)
	}

	// sort the default selection to top if exists
	if opt != nil && opt.Default != nil {
		defaultStr := opt.Field(*opt.Default)
		for i, selection := range selections {
			if defaultStr == selection {
				selections[0], selections[i] = selections[i], selections[0]
				items[0], items[i] = items[i], items[0]
				break
			}
		}
	}

	p := promptui.Select{
		Label:    label,
		Items:    selections,
		Searcher: searcher,
	}

	i, _, err := p.Run()
	if err != nil {
		return *new(T), fmt.Errorf("running select: %w", err)
	}

	if i >= len(items) {
		return *new(T), errors.New("items index out of range")
	}

	return items[i], nil
}

func getCurrentDirName() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %v", err)
	}
	dirName := filepath.Base(dir)
	return sanitizeAppName(dirName), nil
}

// Sanitize the directory name to comply with k8s label rules
func sanitizeAppName(name string) string {
	var builder strings.Builder

	// Remove all characters except alphanumeric, '-', '_', '.'
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			builder.WriteRune(r)
		}
	}

	sanitized := builder.String()
	if sanitized == "" {
		sanitized = defaultAppName
	} else {
		// Trim leading and trailing '-', '_', '.'
		sanitized = strings.Trim(sanitized, "-._")
		// Ensure the length does not exceed 63 characters
		if len(sanitized) > 63 {
			sanitized = sanitized[:63]
		}
	}
	return sanitized
}
