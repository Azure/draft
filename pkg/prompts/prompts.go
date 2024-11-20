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

func RunPromptsFromConfig(draftConfig *config.DraftConfig) error {
	return RunPromptsFromConfigWithSkips(draftConfig)
}

func RunPromptsFromConfigWithSkips(draftConfig *config.DraftConfig) error {
	return RunPromptsFromConfigWithSkipsIO(draftConfig, nil, nil)
}

// RunPromptsFromConfigWithSkipsIO runs the prompts for the given draftConfig
// skipping any variables in varsToSkip or where the BuilderVar.IsPromptDisabled is true.
// If Stdin or Stdout are nil, the default values will be used.
func RunPromptsFromConfigWithSkipsIO(draftConfig *config.DraftConfig, Stdin io.ReadCloser, Stdout io.WriteCloser) error {
	if draftConfig == nil {
		return errors.New("draftConfig is nil")
	}

	for _, variable := range draftConfig.Variables {
		if variable.Value != "" {
			log.Debugf("Skipping prompt for %s", variable.Name)
			continue
		}

		if variable.Default.IsPromptDisabled {
			log.Debugf("Skipping prompt for %s as it has IsPromptDisabled=true", variable.Name)
			noPromptDefaultValue := GetVariableDefaultValue(draftConfig, variable)
			if noPromptDefaultValue == "" {
				return fmt.Errorf("IsPromptDisabled is true for %s but no default value was found", variable.Name)
			}
			log.Debugf("Using default value %s for %s", noPromptDefaultValue, variable.Name)
			variable.Value = noPromptDefaultValue
			continue
		}

		if len(variable.ActiveWhenConstraints) > 0 {
			isVarActive := true
			for _, activeWhen := range variable.ActiveWhenConstraints {
				refVar, err := draftConfig.GetVariable(activeWhen.VariableName)
				if err != nil {
					return fmt.Errorf("unable to get ActiveWhen reference variable: %w", err)
				}

				isConditionTrue, err := draftConfig.CheckActiveWhenConstraint(refVar, activeWhen)
				if err != nil {
					return fmt.Errorf("unable to check ActiveWhen constraint: %w", err)
				}

				if !isConditionTrue {
					isVarActive = false
				}
			}
			if !isVarActive {
				continue
			}
		}

		log.Debugf("constructing prompt for: %s", variable.Name)
		if variable.Type == "bool" {
			input, err := RunBoolPrompt(variable, Stdin, Stdout)
			if err != nil {
				return err
			}
			variable.Value = input
		} else {
			defaultValue := GetVariableDefaultValue(draftConfig, variable)

			stringInput, err := RunDefaultableStringPrompt(defaultValue, variable, nil, Stdin, Stdout)
			if err != nil {
				return err
			}
			variable.Value = stringInput
		}
	}

	return nil
}

// GetVariableDefaultValue returns the default value for a variable, if one is set in variableDefaults from a ReferenceVar or literal Variable.DefaultValue in that order.
func GetVariableDefaultValue(draftConfig *config.DraftConfig, variable *config.BuilderVar) string {
	defaultValue := ""

	if variable.Name == "APPNAME" {
		dirName, err := getCurrentDirNameFunc()
		if err != nil {
			log.Errorf("Error retrieving current directory name: %s", err)
			return defaultAppName
		}
		defaultValue = sanitizeAppName(dirName)
		return defaultValue
	}

	defaultValue = variable.Default.Value
	log.Debugf("setting default value for %s to %s from variable default rule", variable.Name, defaultValue)
	if variable.Default.ReferenceVar != "" {
		if referenceVar, err := draftConfig.GetVariable(variable.Default.ReferenceVar); err != nil {
			log.Errorf("Error getting reference variable %s: %s", variable.Default.ReferenceVar, err)
		} else if referenceVar.Value != "" {
			defaultValue = referenceVar.Value
			log.Debugf("setting default value for %s to %s from referenceVar %s", variable.Name, defaultValue, variable.Default.ReferenceVar)
		}
	}

	return defaultValue
}

func RunBoolPrompt(customPrompt *config.BuilderVar, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
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
func RunDefaultableStringPrompt(defaultValue string, customPrompt *config.BuilderVar, validate func(string) error, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
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

	prompt := &promptui.Prompt{
		Label:    "Please enter " + customPrompt.Description + " (default: " + defaultValue + ")",
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

		searchWords := strings.Split(search, " ")

		for _, word := range searchWords {
			if !strings.Contains(selection, word) {
				return false
			}
		}
		return true
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
		// Ensure the length does not exceed 63 characters
		if len(sanitized) > 63 {
			sanitized = sanitized[:63]
		}
		// Trim leading and trailing '-', '_', '.'
		sanitized = strings.Trim(sanitized, "-._")
	}
	return sanitized
}
