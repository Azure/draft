package prompts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/providers"
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

		log.Debugf("constructing prompt for: %s", variable.Name)
		switch variable.Resource {
		case "ghBranch":
			if input, err := promptGitHubBranch(); err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			} else {
				variable.Value = input
			}
		case "azResourceGroup":
			var input string
			var err error

			switch variable.Name {
			case "ACRRESOURCEGROUP":
				input, err = promptAzResourceGroup("Please select the resource group for your Azure container registry", "")
			case "CLUSTERRESOURCEGROUP":
				var acrResourceGroup *config.BuilderVar
				acrResourceGroup, err = draftConfig.GetVariable("ACRRESOURCEGROUP")
				if err != nil {
					return fmt.Errorf("running prompts from config: %w", err)
				}

				input, err = promptAzResourceGroup("Please select the resource group for your cluster", acrResourceGroup.Value)
			}

			if err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			} else {
				variable.Value = input
			}
		case "azContainerRegistry":
			acrResourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP")
			if err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			}

			if input, err := promptAzContainerRegistry(acrResourceGroup.Value); err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			} else {
				variable.Value = input
			}
		case "containerName":
			azContainerRegistry, err := draftConfig.GetVariable("AZURECONTAINERREGISTRY")
			if err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			}

			if input, err := promptContainerName(azContainerRegistry.Value); err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			} else {
				variable.Value = input
			}
		case "azClusterName":
			clusterResourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
			if err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			}

			if input, err := promptAzClusterName(clusterResourceGroup.Value); err != nil {
				return fmt.Errorf("running prompts from config: %w", err)
			} else {
				variable.Value = input
			}
		default:
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
					fmt.Println(err)
					return err
				}
				variable.Value = stringInput
			}
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
		} else {
			if referenceVar.Value != "" {
				defaultValue = referenceVar.Value
				log.Debugf("setting default value for %s to %s from referenceVar %s", variable.Name, defaultValue, variable.Default.ReferenceVar)
			}
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
		fmt.Println(err)
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
	// Default is the default selection. Don't provide this without a Field function.
	Default *T
}

func Select[T any](label string, items []T, opt *SelectOpt[T]) (T, error) {
	selections := make([]interface{}, len(items))
	for i, item := range items {
		if opt != nil && opt.Field != nil {
			selections[i] = opt.Field(item)
		} else {
			selections[i] = item
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
		// Ensure the length does not exceed 63 characters
		if len(sanitized) > 63 {
			sanitized = sanitized[:63]
		}
		// Trim leading and trailing '-', '_', '.'
		sanitized = strings.Trim(sanitized, "-._")
	}
	return sanitized
}

func promptGitHubBranch() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("prompting for github branch: %w", err)
	}

	currentBranch, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("prompting for github branch: %w", err)
	}

	currentBranchName := currentBranch.Name().Short()

	branches, err := repo.Branches()
	if err != nil {
		return "", fmt.Errorf("prompting for github branch: %w", err)
	}

	var branchNames []string
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		branchNames = append(branchNames, branch.Name().Short())
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("prompting for github branch: %w", err)
	}

	branch, err := Select("Please select the branch for this workflow", branchNames, &SelectOpt[string]{
		Field: func(branchName string) string {
			return branchName
		},
		Default: &currentBranchName,
	})
	if err != nil {
		return "", fmt.Errorf("prompting for github branch: %w", err)
	}

	return branch, nil
}

func promptAzResourceGroup(prompt string, currentResourceGroup string) (string, error) {
	resourceGroups, err := providers.GetAzResourceGroups()
	if err != nil {
		return "", fmt.Errorf("prompting for azure resource group: %w", err)
	}

	var resourceGroup string

	if currentResourceGroup == "" {
		resourceGroup, err = Select(prompt, resourceGroups, nil)
	} else {
		resourceGroup, err = Select(prompt, resourceGroups, &SelectOpt[string]{
			Field: func(resourceGroup string) string {
				return resourceGroup
			},
			Default: &currentResourceGroup,
		})
	}
	if err != nil {
		return "", fmt.Errorf("prompting for azure resource group: %w", err)
	}

	return resourceGroup, nil
}

func promptAzContainerRegistry(resourceGroup string) (string, error) {
	containerRegistries, err := providers.GetAzContainerRegistries(resourceGroup)
	if err != nil {
		return "", fmt.Errorf("prompting for azure container registry: %w", err)
	}

	containerRegistry, err := Select("Please select the container registry for this workflow", containerRegistries, nil)
	if err != nil {
		return "", fmt.Errorf("prompting for azure container registry: %w", err)
	}

	return containerRegistry, nil
}

func promptContainerName(containerRegistry string) (string, error) {
	containerNames, err := providers.GetAzContainerNames(containerRegistry)
	if err != nil {
		return "", fmt.Errorf("prompting for container name: %w", err)
	}

	containerName, err := Select("Please select the container name for this workflow", containerNames, nil)
	if err != nil {
		return "", fmt.Errorf("prompting for container name: %w", err)
	}

	return containerName, nil
}

func promptAzClusterName(resourceGroup string) (string, error) {
	clusters, err := providers.GetAzClusters(resourceGroup)
	if err != nil {
		return "", fmt.Errorf("prompting for azure cluster: %w", err)
	}

	cluster, err := Select("Please select the cluster for this workflow", clusters, nil)
	if err != nil {
		return "", fmt.Errorf("prompting for azure cluster: %w", err)
	}

	return cluster, nil
}
