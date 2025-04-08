package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/draft/pkg/cred"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/providers"
	"github.com/Azure/draft/pkg/spinner"
	"k8s.io/apimachinery/pkg/util/validation"
)

func newSetUpCmd() *cobra.Command {
	sc := &providers.SetUpCmd{}

	// setup-ghCmd represents the setup-gh command
	var cmd = &cobra.Command{
		Use:   "setup-gh",
		Short: "Automates the Github OIDC setup process",
		Long: `This command will automate the Github OIDC setup process by creating an Azure Active Directory 
application and service principle, and will configure that application to trust github.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			gh := providers.NewGhClient()

			azCred, err := cred.GetCred()
			if err != nil {
				return fmt.Errorf("getting credentials: %w", err)
			}
			az, err := providers.NewAzClient(azCred)
			if err != nil {
				return fmt.Errorf("creating azure client: %w", err)
			}
			sc.AzClient = az

			err = fillSetUpConfig(sc, gh, az)
			if err != nil {
				return fmt.Errorf("filling setup config: %w", err)
			}

			s := spinner.CreateSpinner("--> Setting up Github OIDC...")
			s.Start()
			err = runProviderSetUp(ctx, sc, s, gh, az)
			s.Stop()
			if err != nil {
				return err
			}

			log.Info("Draft has successfully set up Github OIDC for your project 😃")
			log.Info("Use 'draft generate-workflow' to generate a Github workflow to build and deploy an application on AKS.")

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.AppName, "app", "a", emptyDefaultFlagValue, "specify the Azure Active Directory application name")
	f.StringVarP(&sc.SubscriptionID, "subscription-id", "s", emptyDefaultFlagValue, "specify the Azure subscription ID")
	f.StringVarP(&sc.ResourceGroupName, "resource-group", "r", emptyDefaultFlagValue, "specify the Azure resource group name")
	f.StringVarP(&sc.Repo, "gh-repo", "g", emptyDefaultFlagValue, "specify the github repository link")
	sc.Provider = provider
	return cmd
}

func fillSetUpConfig(sc *providers.SetUpCmd, gh providers.GhClient, az providers.AzClientInterface) error {
	if sc.TenantId == "" {
		tenandId, err := providers.PromptTenantId(sc.AzClient, context.Background())
		if err != nil {
			return fmt.Errorf("prompting tenant ID: %w", err)
		}
		sc.TenantId = tenandId
	}

	if sc.AppName == "" {
		// Set the application name; default is the current directory name plus "-workflow".

		// get the current directory name
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		defaultAppName := fmt.Sprintf("%s-workflow", filepath.Base(currentDir))
		defaultAppName, err = ToValidAppName(defaultAppName)
		if err != nil {
			log.Debugf("unable to convert default app name %q to a valid name: %v", defaultAppName, err)
			log.Debugf("using default app name %q", defaultAppName)
			defaultAppName = "my-workflow"
		}

		appName, err := PromptAppName(sc.AzClient, defaultAppName)
		if err != nil {
			return fmt.Errorf("prompting app name: %w", err)
		}
		sc.AppName = appName
	}

	if sc.SubscriptionID == "" {
		if strings.ToLower(sc.Provider) == "azure" {
			currentSub, err := az.GetCurrentAzSubscriptionLabel()
			if err != nil {
				return fmt.Errorf("getting current subscription ID: %w", err)
			}

			subLabels, err := az.GetAzSubscriptionLabels()
			if err != nil {
				return fmt.Errorf("getting subscription labels: %w", err)
			}

			sc.SubscriptionID, err = getAzSubscriptionId(subLabels, currentSub)
			if err != nil {
				return fmt.Errorf("getting subscription ID: %w", err)
			}
		} else {
			sc.SubscriptionID = getSubscriptionID()
		}
	}

	if sc.ResourceGroupName == "" {
		rg, err := PromptResourceGroup(sc.AzClient, sc.SubscriptionID)
		if err != nil {
			return fmt.Errorf("getting resource group: %w", err)
		}
		sc.ResourceGroupName = *rg.Name
	}

	if sc.Repo == "" {
		repo, err := PromptGitHubRepoWithOwner(gh)
		if err != nil {
			return fmt.Errorf("failed to prompt for GitHub repository: %w", err)
		}
		if repo == "" {
			return errors.New("github repo cannot be empty")
		}
		sc.Repo = repo
	}

	return nil
}

func ToValidAppName(name string) (string, error) {
	// replace all underscores with hyphens
	cleanedName := strings.ReplaceAll(name, "_", "-")
	// replace all spaces with hyphens
	cleanedName = strings.ReplaceAll(cleanedName, " ", "-")

	// remove leading non-alphanumeric characters
	for i, r := range cleanedName {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			cleanedName = cleanedName[i:]
			break
		}
	}

	// remove trailing non-alphanumeric characters
	for i := len(cleanedName) - 1; i >= 0; i-- {
		r := rune(cleanedName[i])
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			cleanedName = cleanedName[:i+1]
			break
		}
	}

	// remove all characters except alphanumeric, '-', '.'
	var builder strings.Builder
	for _, r := range cleanedName {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			builder.WriteRune(r)
		}
	}

	// lowercase the name
	cleanedName = strings.ToLower(builder.String())
	if err := ValidateAppName(cleanedName); err != nil {
		return "", fmt.Errorf("app name '%s' could not be converted to a valid name: %w", name, err)
	}
	return cleanedName, nil
}

func runProviderSetUp(ctx context.Context, sc *providers.SetUpCmd, s spinner.Spinner, gh providers.GhClient, az providers.AzClientInterface) error {
	provider := strings.ToLower(sc.Provider)
	if provider == "azure" {
		// call azure provider logic
		return providers.InitiateAzureOIDCFlow(ctx, sc, s, gh, az)

	} else {
		// call logic for user-submitted provider
		fmt.Printf("The provider is %v\n", sc.Provider)
	}

	return nil
}

func ValidateAppName(name string) error {
	errors := validation.IsDNS1123Label(name)
	if len(errors) > 0 {
		return fmt.Errorf("invalid app name: %s", strings.Join(errors, ", "))
	}
	return nil
}

func PromptAppName(az providers.AzClientInterface, defaultAppName string) (string, error) {
	appNamePrompt := &promptui.Prompt{
		Label:    "Enter app registration name",
		Validate: ValidateAppName,
		Default:  defaultAppName,
	}
	appName, err := appNamePrompt.Run()

	if err != nil {
		return "", err
	}

	if az.AzAppExists(appName) {
		confirmAppExistsPrompt := promptui.Prompt{
			Label:     "An app with this name already exists. Would you like to use it?",
			IsConfirm: true,
		}
		_, err := confirmAppExistsPrompt.Run()
		if err != nil {
			return PromptAppName(az, defaultAppName)
		}
	} else {
		log.Debugf("App %q does not exist, will be created", appName)
	}

	return appName, nil
}

func getSubscriptionID() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("invalid subscription id")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter subscription ID",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return err.Error()
	}

	return result
}

func PromptResourceGroup(az providers.AzClientInterface, subscriptionID string) (armresources.ResourceGroup, error) {
	var rg armresources.ResourceGroup
	log.Println("Fetching resource groups...")
	rgs, err := az.ListResourceGroups(context.Background(), subscriptionID)
	if err != nil {
		return rg, fmt.Errorf("listing resource groups: %w", err)
	}

	rg, err = prompts.Select("Please choose the resource group you would like to use", rgs, &prompts.SelectOpt[armresources.ResourceGroup]{
		Field: func(rg armresources.ResourceGroup) string {
			return *rg.Name + " (" + *rg.Location + ")"
		},
	})
	if err != nil {
		return rg, fmt.Errorf("selecting resource group: %w", err)
	}

	return rg, nil
}

func PromptGitHubRepoWithOwner(gh providers.GhClient) (string, error) {
	defaultRepoNameWithOwner, err := gh.GetRepoNameWithOwner()
	if err != nil {
		return "", err
	}
	log.Println("Prompting for github repo with owner name...")
	repoPrompt := promptui.Prompt{
		Label: "Enter github organization and repo organization and repoName",
		Validate: func(input string) error {
			if !strings.Contains(input, "/") {
				return errors.New("github repo cannot be empty")
			}
			return nil
		},
		Default: defaultRepoNameWithOwner,
	}

	repo, err := repoPrompt.Run()
	if err != nil {
		return "", fmt.Errorf("running repo name with owner prompt: %w", err)
	}

	log.Debug("Validating github repo...")
	if err := gh.IsValidGhRepo(repo); err != nil {
		confirmMissingRepoPrompt := promptui.Prompt{
			Label:     "Unable to confirm this repo exists. Do you want to proceed anyway?",
			IsConfirm: true,
		}
		_, err := confirmMissingRepoPrompt.Run()
		if err != nil {
			return PromptGitHubRepoWithOwner(gh)
		}
	} else {
		log.Debugf("Github repo %q is valid", repo)
	}
	return repo, nil
}

func getCloudProvider() string {
	selection := &promptui.Select{
		Label: "What cloud provider would you like to use?",
		Items: []string{"azure"},
	}

	_, selectResponse, err := selection.Run()
	if err != nil {
		return err.Error()
	}

	return selectResponse
}

func getAzSubscriptionId(subLabels []providers.SubLabel, currentSub providers.SubLabel) (string, error) {
	subLabel, err := prompts.Select("Please choose the subscription ID you would like to use", subLabels, &prompts.SelectOpt[providers.SubLabel]{
		Field: func(subLabel providers.SubLabel) string {
			return subLabel.Name + " (" + subLabel.ID + ")"
		},
		Default: &currentSub,
	})
	if err != nil {
		return "", fmt.Errorf("selecting subscription ID: %w", err)
	}

	return subLabel.ID, nil
}

func init() {
	rootCmd.AddCommand(newSetUpCmd())
}
