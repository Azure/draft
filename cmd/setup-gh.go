package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/Azure/draft/pkg/cred"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/providers"
	"github.com/Azure/draft/pkg/spinner"
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

			azCred, err := cred.GetCred()
			if err != nil {
				return fmt.Errorf("getting credentials: %w", err)
			}

			client, err := armsubscription.NewTenantsClient(azCred, nil)
			if err != nil {
				return fmt.Errorf("creating tenants client: %w", err)
			}

			sc.AzClient.AzTenantClient = client

			err = fillSetUpConfig(sc)
			if err != nil {
				return fmt.Errorf("filling setup config: %w", err)
			}

			s := spinner.CreateSpinner("--> Setting up Github OIDC...")
			s.Start()
			err = runProviderSetUp(ctx, sc, s)
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

func fillSetUpConfig(sc *providers.SetUpCmd) error {
	if sc.AppName == "" {
		sc.AppName = getAppName()
	}

	if sc.SubscriptionID == "" {
		if strings.ToLower(sc.Provider) == "azure" {
			currentSub, err := providers.GetCurrentAzSubscriptionLabel()
			if err != nil {
				return fmt.Errorf("getting current subscription ID: %w", err)
			}

			subLabels, err := providers.GetAzSubscriptionLabels()
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
		sc.ResourceGroupName = getResourceGroup()
	}

	if sc.Repo == "" {
		sc.Repo = getGhRepo()
	}

	return nil
}

func runProviderSetUp(ctx context.Context, sc *providers.SetUpCmd, s spinner.Spinner) error {
	provider := strings.ToLower(sc.Provider)
	if provider == "azure" {
		// call azure provider logic
		return providers.InitiateAzureOIDCFlow(ctx, sc, s)

	} else {
		// call logic for user-submitted provider
		fmt.Printf("The provider is %v\n", sc.Provider)
	}

	return nil
}

func getAppName() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("Invalid app name")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter app registration name",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return err.Error()
	}

	return result
}

func getSubscriptionID() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("Invalid subscription id")
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

func getResourceGroup() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("Invalid resource group name")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter resource group name",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return err.Error()
	}

	return result
}

func getGhRepo() string {
	validate := func(input string) error {
		if !strings.Contains(input, "/") {
			return errors.New("Github repo cannot be empty")
		}

		return nil
	}

	repoPrompt := promptui.Prompt{
		Label:    "Enter github organization and repo (organization/repoName)",
		Validate: validate,
	}

	repo, err := repoPrompt.Run()
	if err != nil {
		return err.Error()
	}

	return repo
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
