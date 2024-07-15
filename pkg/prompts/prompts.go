package prompts

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/osutil"
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
	action := []string{"create", "existing"}

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

		var choice string
		var input string
		var err error

		log.Debugf("constructing prompt for: %s", variable.Name)
		switch variable.Resource {
		case "ghBranch":
			if choice, err = Select("Please choose if you'd like to create a new branch or use an existing branch", action, nil); err != nil {
				return fmt.Errorf("failed to select an action: %w", err)
			}

			switch choice {
			case "create":
				if input, err = createGitHubBranch(variable); err != nil {
					return fmt.Errorf("failed to create GitHub branch: %w", err)
				}
			case "existing":
				if input, err = promptGitHubBranch(); err != nil {
					return fmt.Errorf("failed to prompt for GitHub branch: %w", err)
				}
			}
		case "azResourceGroup":
			switch variable.Name {
			case "ACRRESOURCEGROUP":
				if choice, err = Select("Please choose if you'd like to create a new resource group or use an existing resource group for your Azure container registry", action, nil); err != nil {
					return fmt.Errorf("failed to select an action: %w", err)
				}

				switch choice {
				case "create":
					if input, err = createAzResourceGroup(variable); err != nil {
						return fmt.Errorf("failed to create Azure resource group: %w", err)
					}
				case "existing":
					if input, err = promptAzResourceGroup("Please select the resource group for your Azure container registry", ""); err != nil {
						return fmt.Errorf("failed to prompt for Azure resource group: %w", err)
					}
				}
			case "CLUSTERRESOURCEGROUP":
				if choice, err = Select("Please choose if you'd like to create a new resource group or use an existing resource group for your Azure cluster", action, nil); err != nil {
					return fmt.Errorf("failed to select an action: %w", err)
				}

				switch choice {
				case "create":
					if input, err = createAzResourceGroup(variable); err != nil {
						return fmt.Errorf("failed to create Azure resource group: %w", err)
					}
				case "existing":
					if acrResourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP"); err != nil {
						return fmt.Errorf("failed to get variable: %w", err)
					} else if input, err = promptAzResourceGroup("Please select the resource group for your cluster", acrResourceGroup.Value); err != nil {
						return fmt.Errorf("failed to prompt for Azure resource group: %w", err)
					}
				}
			}
		case "azContainerRegistry":
			if choice, err = Select("Please choose if you'd like to create a new Azure container registry or use an existing Azure container registry", action, nil); err != nil {
				return fmt.Errorf("failed to select an action: %w", err)
			}

			acrResourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP")
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			switch choice {
			case "create":
				if input, err = createAzContainerRegistry(variable, acrResourceGroup.Value); err != nil {
					return fmt.Errorf("failed to create Azure container registry: %w", err)
				}
			case "existing":
				if input, err = promptAzContainerRegistry(acrResourceGroup.Value); err != nil {
					return fmt.Errorf("failed to prompt for azure container registry: %w", err)
				}
			}
		case "azClusterName":
			if choice, err = Select("Please choose if you'd like to create a new Azure cluster or use an existing Azure cluster", action, nil); err != nil {
				return fmt.Errorf("failed to select an action: %w", err)
			}

			clusterResourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			switch choice {
			case "create":
				if input, err = createAzCluster(variable, clusterResourceGroup.Value); err != nil {
					return fmt.Errorf("failed to create Azure cluster: %w", err)
				}
			case "existing":
				if input, err = promptAzClusterName(clusterResourceGroup.Value); err != nil {
					return fmt.Errorf("failed to prompt for Azure cluster name: %w", err)
				}
			}
		case "azNamespace":
			if choice, err = Select("Please choose if you'd like to create a new Azure namespace or use an existing Azure namespace", action, nil); err != nil {
				return fmt.Errorf("failed to select an action: %w", err)
			}

			clusterResourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			clusterName, err := draftConfig.GetVariable("CLUSTERNAME")
			if err != nil {
				return fmt.Errorf("failed to get variable: %w", err)
			}

			switch choice {
			case "create":
				if input, err = createAzNamespace(variable, clusterResourceGroup.Value, clusterName.Value); err != nil {
					return fmt.Errorf("failed to create Azure namespace: %w", err)
				}
			case "existing":
				if input, err = promptAzNamespace(clusterResourceGroup.Value, clusterName.Value); err != nil {
					return fmt.Errorf("failed to prompt for Azure namespace: %w", err)
				}
			}
		default:
			if variable.Type == "bool" {
				if input, err = RunBoolPrompt(variable, Stdin, Stdout); err != nil {
					return fmt.Errorf("failed to run bool prompt: %w", err)
				}
			} else {
				defaultValue := GetVariableDefaultValue(draftConfig, variable)

				if input, err = RunDefaultableStringPrompt(defaultValue, variable, nil, Stdin, Stdout); err != nil {
					return fmt.Errorf("failed to run defaultable string prompt: %w", err)
				}
			}
		}

		variable.Value = input
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

func validateAzResourceGroup(resourceGroup string) error {
	resourceGroupRegEx := regexp.MustCompile(`^[a-z0-9._-]+$`)

	if len(resourceGroup) < 1 || len(resourceGroup) > 90 || !resourceGroupRegEx.MatchString(resourceGroup) || resourceGroup[len(resourceGroup)-1] == '.' {
		return fmt.Errorf("invalid resource group name:\n1. Between 1 and 90 characters long\n2. Only include lowercase alphanumeric characters, periods, underscores, and hyphens\n3. Can't end with a period")
	}

	return nil
}

func validateAzRepositoryName(repositoryName string) error {
	repositoryNameRegEx := regexp.MustCompile(`^[a-z0-9._/-]+$`)

	if !repositoryNameRegEx.MatchString(repositoryName) {
		return fmt.Errorf("invalid repository name:\nRepository names can only include lowercase alphanumeric characters, periods, hyphens, underscores, and forward slashes")
	}

	return nil
}

func validateAzContainerRegistry(containerRegistry string) error {
	acrRegEx := regexp.MustCompile(`^[a-z0-9]+$`)

	if len(containerRegistry) < 5 || len(containerRegistry) > 50 || !acrRegEx.MatchString(containerRegistry) {
		return fmt.Errorf("invalid container registry name:\n1. Between 5 and 50 characters long\n2. Only include lowercase alphanumeric characters\n3. Can't end with a period")
	}

	return nil
}

func validateAzClusterName(clusterName string) error {
	clusterNameRegEx := regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

	if len(clusterName) > 253 || !clusterNameRegEx.MatchString(clusterName) {
		return fmt.Errorf("invalid cluster name:\n1. Between 1 and 253 characters long\n2. Only include lowercase alphanumeric characters, periods, and hyphens\n3. Must start and end with a lowercase alphanumeric character")
	}

	return nil
}

func validateAzNamespace(namespace string) error {
	namespaceRegEx := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

	if len(namespace) > 63 || !namespaceRegEx.MatchString(namespace) {
		return fmt.Errorf("invalid namespace name:\n1. Between 1 and 63 characters long\n2. Only include lowercase alphanumeric characters and hyphens\n3. Must start and end with a lowercase alphanumeric character")
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
		} else if customPrompt.Resource == "azResourceGroup" {
			if err := validateAzResourceGroup(input); err != nil {
				return err
			}
		} else if customPrompt.Resource == "azContainerRegistry" {
			if err := validateAzContainerRegistry(input); err != nil {
				return err
			}
		} else if customPrompt.Resource == "azRepositoryName" {
			if err := validateAzRepositoryName(input); err != nil {
				return err
			}
		} else if customPrompt.Resource == "azClusterName" {
			if err := validateAzClusterName(input); err != nil {
				return err
			}
		} else if customPrompt.Resource == "azNamespace" {
			if err := validateAzNamespace(input); err != nil {
				return err
			}
		} else if customPrompt.Resource == "path" {
			if err := osutil.CheckPath(input); err != nil {
				return err
			}
		} else {
			if err := validate(input); err != nil {
				return err
			}
		}
		return nil
	}

	var promptLabel string

	if defaultValue == "" {
		promptLabel = "Please enter " + customPrompt.Description
	} else {
		promptLabel = "Please enter " + customPrompt.Description + " (default: " + defaultValue + ")"
	}

	prompt := &promptui.Prompt{
		Label:    promptLabel,
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

func createGitHubBranch(branchVar *config.BuilderVar) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	branchName, err := RunDefaultableStringPrompt("", branchVar, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create branch: %w", err)
	}

	return branchName, nil
}

func createAzResourceGroup(azrgVar *config.BuilderVar) (string, error) {
	resourceGroup, err := RunDefaultableStringPrompt("", azrgVar, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	if locations, err := providers.GetAzLocations(); err != nil {
		return "", fmt.Errorf("failed to get Azure locations: %w", err)
	} else if location, err := Select("Please select the location for this resource group", locations, nil); err != nil {
		return "", fmt.Errorf("failed to select a location: %w", err)
	} else if err := providers.CreateAzResourceGroup(resourceGroup, location); err != nil {
		return "", fmt.Errorf("failed to create Azure resource group: %w", err)
	}

	return resourceGroup, nil
}

func createAzContainerRegistry(acr *config.BuilderVar, resourceGroup string) (string, error) {
	containerRegistry, err := RunDefaultableStringPrompt("", acr, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	skus := []string{"Basic", "Standard", "Premium"}

	if sku, err := Select("Please select the SKU for this Azure container registry", skus, nil); err != nil {
		return "", fmt.Errorf("failed to select a SKU: %w", err)
	} else if err := providers.CreateAzContainerRegistry(containerRegistry, resourceGroup, sku); err != nil {
		return "", fmt.Errorf("failed to create Azure container registry: %w", err)
	}

	return containerRegistry, nil
}

func createAzCluster(azClusterVar *config.BuilderVar, clusterResourceGroup string) (string, error) {
	clusterName, err := RunDefaultableStringPrompt("", azClusterVar, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	settings := []string{"public", "private"}

	if setting, err := Select("Please select the privacy setting for this Azure cluster", settings, nil); err != nil {
		return "", fmt.Errorf("failed to select a privacy setting: %w", err)
	} else if err = providers.CreateAzCluster(clusterName, clusterResourceGroup, setting); err != nil {
		return "", fmt.Errorf("failed to create Azure cluster: %w", err)
	}

	return clusterName, nil
}

func createAzNamespace(namespaceVar *config.BuilderVar, clusterResourceGroup, clusterName string) (string, error) {
	namespace, err := RunDefaultableStringPrompt("", namespaceVar, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	if err := providers.CreateAzNamespace(namespace, clusterResourceGroup, clusterName); err != nil {
		return "", fmt.Errorf("failed to create Azure namespace: %w", err)
	}

	return namespace, nil
}

func promptGitHubBranch() (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	currentBranch, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve current branch: %w", err)
	}

	currentBranchName := currentBranch.Name().Short()

	branches, err := repo.Branches()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve branches: %w", err)
	}

	var branchNames []string
	err = branches.ForEach(func(branch *plumbing.Reference) error {
		branchNames = append(branchNames, branch.Name().Short())
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to create branchNames: %w", err)
	}

	branch, err := Select("Please select the branch for this workflow", branchNames, &SelectOpt[string]{
		Field: func(branchName string) string {
			return branchName
		},
		Default: &currentBranchName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select a branch: %w", err)
	}

	return branch, nil
}

func promptAzResourceGroup(prompt string, currentResourceGroup string) (string, error) {
	resourceGroups, err := providers.GetAzResourceGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get Azure resource group: %w", err)
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
		return "", fmt.Errorf("failed to select a resource group: %w", err)
	}

	return resourceGroup, nil
}

func promptAzContainerRegistry(resourceGroup string) (string, error) {
	containerRegistries, err := providers.GetAzContainerRegistries(resourceGroup)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure container registries: %w", err)
	}

	containerRegistry, err := Select("Please select the container registry for this workflow", containerRegistries, nil)
	if err != nil {
		return "", fmt.Errorf("failed to select a container registry: %w", err)
	}

	return containerRegistry, nil
}

func promptAzClusterName(clusterResourceGroup string) (string, error) {
	clusters, err := providers.GetAzClusters(clusterResourceGroup)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure clusters: %w", err)
	}

	cluster, err := Select("Please select the cluster for this workflow", clusters, nil)
	if err != nil {
		return "", fmt.Errorf("failed to select a cluster: %w", err)
	}

	return cluster, nil
}

func promptAzNamespace(clusterResourceGroup string, clusterName string) (string, error) {
	namespaces, err := providers.GetAzNamespaces(clusterResourceGroup, clusterName)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure namespaces: %w", err)
	}

	namespace, err := Select("Please select the namespace for this workflow", namespaces, nil)
	if err != nil {
		return "", fmt.Errorf("failed to select a namespace: %w", err)
	}

	return namespace, nil
}
