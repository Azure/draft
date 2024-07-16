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
	"github.com/Azure/draft/pkg/providers"
)

const defaultAppName = "my-app"

// map for resource types to their respective prompting functions
type resourceFunc func(*config.DraftConfig, *config.BuilderVar) (string, error)

var resourceFuncMap = map[string]resourceFunc{
	"ghBranch":            promptGitHubBranch,
	"azResourceGroup":     promptAzResourceGroup,
	"azContainerRegistry": promptAzContainerRegistry,
	"azClusterName":       promptAzClusterName,
	"azNamespace":         promptAzNamespace,
}

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

		var input string
		var err error

		log.Debugf("constructing prompt for: %s", variable.Name)
		if resourceFunc, ok := resourceFuncMap[variable.Resource]; ok {
			if input, err = resourceFunc(draftConfig, variable); err != nil {
				return fmt.Errorf("failed to prompt for %s: %w", variable.Name, err)
			}
		} else {
			if variable.Type == "bool" {
				if input, err = RunBoolPrompt(variable, Stdin, Stdout); err != nil {
					return fmt.Errorf("failed to run bool prompt: %w", err)
				}
			} else {
				defaultValue := GetVariableDefaultValue(draftConfig, variable)

				if variable.Name == "APPNAME" {
					input, err = RunDefaultableStringPrompt(defaultValue, variable, appNameValidator, Stdin, Stdout)
				} else if variable.Resource == "azRepositoryName" {
					input, err = RunDefaultableStringPrompt(defaultValue, variable, validateAzRepositoryName, Stdin, Stdout)
				} else {
					input, err = RunDefaultableStringPrompt(defaultValue, variable, nil, Stdin, Stdout)
				}

				if err != nil {
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
	resourceGroupRegEx := regexp.MustCompile(`^[\p{L}\p{N}_\-\.\(\)]+$`)

	if len(resourceGroup) == 0 || len(resourceGroup) > 90 {
		return fmt.Errorf("resource group must be between 1 and 90 characters long")
	}

	if !resourceGroupRegEx.MatchString(resourceGroup) {
		return fmt.Errorf("resource group names can only include alphanumeric characters, periods, underscores, hyphens, and parantheses")
	}

	if resourceGroup[len(resourceGroup)-1] == '.' {
		return fmt.Errorf("resource group names can't end with a period")
	}

	return nil
}

func validateAzRepositoryName(repositoryName string) error {
	repositoryNameRegEx := regexp.MustCompile(`^[a-z0-9._/-]+$`)

	if len(repositoryName) == 0 || !repositoryNameRegEx.MatchString(repositoryName) {
		return fmt.Errorf("repository names can only include lowercase alphanumeric characters, periods, hyphens, underscores, and forward slashes")
	}

	return nil
}

func validateAzContainerRegistry(containerRegistry string) error {
	acrRegEx := regexp.MustCompile(`^[a-z0-9]+$`)

	if len(containerRegistry) < 5 || len(containerRegistry) > 50 {
		return fmt.Errorf("container registry name must be between 5 and 50 characters long")
	}

	if !acrRegEx.MatchString(containerRegistry) {
		return fmt.Errorf("container registry names can only include lowercase alphanumeric characters")
	}

	return nil
}

func validateAzClusterName(clusterName string) error {
	clusterNameRegEx := regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

	if len(clusterName) == 0 || len(clusterName) > 253 {
		return fmt.Errorf("cluster names must be between 1 and 253 characters long")
	}

	if !clusterNameRegEx.MatchString(clusterName) {
		return fmt.Errorf("cluster names can only include lowercase alphanumeric characters, periods, and hyphens and must start and end with a lowercase alphanumeric")
	}

	return nil
}

func validateAzNamespace(namespace string) error {
	namespaceRegEx := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

	if len(namespace) == 0 || len(namespace) > 63 {
		return fmt.Errorf("namespaces must be between 1 and 63 characters long")
	}

	if !namespaceRegEx.MatchString(namespace) {
		return fmt.Errorf("namespaces can only include lowercase alphanumeric characters and hyphens and must start and end with a lowercase alphanumeric")
	}

	return nil
}

// RunDefaultableStringPrompt runs a prompt for a string variable, returning the user string input for the prompt
func RunDefaultableStringPrompt(defaultValue string, customPrompt *config.BuilderVar, validate func(string) error, Stdin io.ReadCloser, Stdout io.WriteCloser) (string, error) {
	var prompt *promptui.Prompt

	if defaultValue == "" {
		if validate == nil {
			validate = NoBlankStringValidator
		}

		prompt = &promptui.Prompt{
			Label:    "Please enter " + customPrompt.Description,
			Validate: validate,
			Stdin:    Stdin,
			Stdout:   Stdout,
		}
	} else {
		prompt = &promptui.Prompt{
			Label:    "Please input " + customPrompt.Description + " (press enter to use the default selection: " + defaultValue + ")",
			Validate: validate,
			Stdin:    Stdin,
			Stdout:   Stdout,
		}
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
	DraftConfig *config.DraftConfig
	BuilderVar  *config.BuilderVar
	// Create is a function that will be called when the user selects the create option.
	Create func(*config.DraftConfig, *config.BuilderVar) (T, error)
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

	if len(selections) > 0 {
		if _, ok := selections[0].(string); !ok {
			return *new(T), errors.New("selections must be of type string or use opt.Field")
		}
	}

	if len(selections) == 0 && opt.Create == nil {
		return *new(T), errors.New("no selection options and no create function")
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

	startIdx := 0

	if opt != nil && opt.Create != nil {
		createOpt := []interface{}{"Create New Item"}
		selections = append(createOpt, selections...)
		startIdx = 1
	}

	p := promptui.Select{
		CursorPos: startIdx,
		Label:     label,
		Items:     selections,
		Searcher:  searcher,
	}

	i, _, err := p.Run()
	if err != nil {
		return *new(T), fmt.Errorf("running select: %w", err)
	}

	if i >= len(items) && opt.Create == nil {
		return *new(T), errors.New("items index out of range")
	}

	if opt != nil && opt.Create != nil && i == 0 {
		if opt.DraftConfig == nil {
			return *new(T), errors.New("Create() provided but draft config is nil")
		} else if opt.BuilderVar == nil {
			return *new(T), errors.New("Create() provided but builder var is nil")
		} else {
			return opt.Create(opt.DraftConfig, opt.BuilderVar)
		}
	} else if opt != nil && opt.Create != nil {
		// create option only exists in selections slice, not items slice
		return items[i-1], nil
	} else {
		return items[i], nil
	}
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

func createGitHubBranch(draftConfig *config.DraftConfig, ghBranch *config.BuilderVar) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	branchName, err := RunDefaultableStringPrompt("", ghBranch, nil, nil, nil)
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

func promptGitHubBranch(draftConfig *config.DraftConfig, ghBranch *config.BuilderVar) (string, error) {
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
		DraftConfig: draftConfig,
		BuilderVar:  ghBranch,
		Create:      createGitHubBranch,
		Field: func(branchName string) string {
			return branchName
		},
		Default: &currentBranchName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select a branch: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return "", fmt.Errorf("failed to checkout branch: %w", err)
	}

	return branch, nil
}

func createAzResourceGroup(draftConfig *config.DraftConfig, resourceGroup *config.BuilderVar) (string, error) {
	resourceGroupVal, err := RunDefaultableStringPrompt("", resourceGroup, validateAzResourceGroup, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	if locations, err := providers.GetAzLocations(); err != nil {
		return "", fmt.Errorf("failed to get Azure locations: %w", err)
	} else if location, err := Select("Please select the location for this resource group", locations, nil); err != nil {
		return "", fmt.Errorf("failed to select a location: %w", err)
	} else if err := providers.CreateAzResourceGroup(resourceGroupVal, location); err != nil {
		return "", fmt.Errorf("failed to create Azure resource group: %w", err)
	}

	return resourceGroupVal, nil
}

func promptAzResourceGroup(draftConfig *config.DraftConfig, resourceGroup *config.BuilderVar) (string, error) {
	resourceGroups, err := providers.GetAzResourceGroups()
	if err != nil {
		return "", fmt.Errorf("failed to get Azure resource group: %w", err)
	}

	var currentResourceGroup string

	switch resourceGroup.Name {
	case "ACRRESOURCEGROUP":
		currentResourceGroup = ""
	case "CLUSTERRESOURCEGROUP":
		if acrResourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP"); err != nil {
			return "", fmt.Errorf("failed to get variable: %w", err)
		} else {
			currentResourceGroup = acrResourceGroup.Value
		}
	}

	var resourceGroupVal string

	if currentResourceGroup == "" {
		resourceGroupVal, err = Select("Please select the resource group for your Azure container registry", resourceGroups, &SelectOpt[string]{
			DraftConfig: draftConfig,
			BuilderVar:  resourceGroup,
			Create:      createAzResourceGroup,
		})
	} else {
		resourceGroupVal, err = Select("Please select the resource group for your cluster", resourceGroups, &SelectOpt[string]{
			DraftConfig: draftConfig,
			BuilderVar:  resourceGroup,
			Create:      createAzResourceGroup,
			Field: func(resourceGroup string) string {
				return resourceGroup
			},
			Default: &currentResourceGroup,
		})
	}
	if err != nil {
		return "", fmt.Errorf("failed to select a resource group: %w", err)
	}

	return resourceGroupVal, nil
}

func createAzContainerRegistry(draftConfig *config.DraftConfig, acr *config.BuilderVar) (string, error) {
	containerRegistry, err := RunDefaultableStringPrompt("", acr, validateAzContainerRegistry, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	skus := []string{"Basic", "Standard", "Premium"}

	if sku, err := Select("Please select the SKU for this Azure container registry", skus, nil); err != nil {
		return "", fmt.Errorf("failed to select a SKU: %w", err)
	} else if resourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP"); err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	} else if err := providers.CreateAzContainerRegistry(containerRegistry, resourceGroup.Value, sku); err != nil {
		return "", fmt.Errorf("failed to create Azure container registry: %w", err)
	}

	return containerRegistry, nil
}

func promptAzContainerRegistry(draftConfig *config.DraftConfig, acr *config.BuilderVar) (string, error) {
	resourceGroup, err := draftConfig.GetVariable("ACRRESOURCEGROUP")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	containerRegistries, err := providers.GetAzContainerRegistries(resourceGroup.Value)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure container registries: %w", err)
	}

	containerRegistry, err := Select("Please select the container registry for this workflow", containerRegistries, &SelectOpt[string]{
		DraftConfig: draftConfig,
		BuilderVar:  acr,
		Create:      createAzContainerRegistry,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select a container registry: %w", err)
	}

	return containerRegistry, nil
}

func createAzCluster(draftConfig *config.DraftConfig, clusterName *config.BuilderVar) (string, error) {
	clusterNameVal, err := RunDefaultableStringPrompt("", clusterName, validateAzClusterName, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	settings := []string{"public", "private"}

	if setting, err := Select("Please select the privacy setting for this Azure cluster", settings, nil); err != nil {
		return "", fmt.Errorf("failed to select a privacy setting: %w", err)
	} else if resourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP"); err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	} else if err = providers.CreateAzCluster(clusterNameVal, resourceGroup.Value, setting); err != nil {
		return "", fmt.Errorf("failed to create Azure cluster: %w", err)
	}

	return clusterNameVal, nil
}

func promptAzClusterName(draftConfig *config.DraftConfig, clusterName *config.BuilderVar) (string, error) {
	resourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	clusters, err := providers.GetAzClusters(resourceGroup.Value)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure clusters: %w", err)
	}

	cluster, err := Select("Please select the cluster for this workflow", clusters, &SelectOpt[string]{
		DraftConfig: draftConfig,
		BuilderVar:  clusterName,
		Create:      createAzCluster,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select a cluster: %w", err)
	}

	return cluster, nil
}

func createAzNamespace(draftConfig *config.DraftConfig, namespaceVar *config.BuilderVar) (string, error) {
	resourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	clusterName, err := draftConfig.GetVariable("CLUSTERNAME")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	namespace, err := RunDefaultableStringPrompt("", namespaceVar, validateAzNamespace, nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to run defaultable string prompt: %w", err)
	}

	if err := providers.CreateAzNamespace(namespace, resourceGroup.Value, clusterName.Value); err != nil {
		return "", fmt.Errorf("failed to create Azure namespace: %w", err)
	}

	return namespace, nil
}

func promptAzNamespace(draftConfig *config.DraftConfig, namespace *config.BuilderVar) (string, error) {
	resourceGroup, err := draftConfig.GetVariable("CLUSTERRESOURCEGROUP")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	clusterName, err := draftConfig.GetVariable("CLUSTERNAME")
	if err != nil {
		return "", fmt.Errorf("failed to get variable: %w", err)
	}

	namespaces, err := providers.GetAzNamespaces(resourceGroup.Value, clusterName.Value)
	if err != nil {
		return "", fmt.Errorf("failed to get Azure namespaces: %w", err)
	}

	namespaceVal, err := Select("Please select the namespace for this workflow", namespaces, &SelectOpt[string]{
		DraftConfig: draftConfig,
		BuilderVar:  namespace,
		Create:      createAzNamespace,
	})
	if err != nil {
		return "", fmt.Errorf("failed to select a namespace: %w", err)
	}

	return namespaceVal, nil
}
