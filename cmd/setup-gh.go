package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/Azure/draft/pkg/cred"
	"github.com/manifoldco/promptui"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"

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

			graphClient, err := createGraphClient(azCred)
			if err != nil {
				return fmt.Errorf("getting client: %w", err)
			}

			sc.AzClient.GraphClient = graphClient

			fillSetUpConfig(sc)

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

func createGraphClient(azCred *azidentity.DefaultAzureCredential) (providers.GraphClient, error) {
	client, err := msgraph.NewGraphServiceClientWithCredentials(azCred, []string{cloud.AzurePublic.Services[cloud.ResourceManager].Endpoint + "/.default"})
	if err != nil {
		return nil, fmt.Errorf("creating graph service client: %w", err)
	}
	return &providers.GraphServiceClient{Client: client}, nil
}

func fillSetUpConfig(sc *providers.SetUpCmd) {
	if sc.AppName == "" {
		sc.AppName = getAppName()
	}

	if sc.SubscriptionID == "" {
		if strings.ToLower(sc.Provider) == "azure" {
			currentSub := providers.GetCurrentAzSubscriptionId()
			sc.SubscriptionID = GetAzSubscriptionId(currentSub)
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

func GetAzSubscriptionId(subIds []string) string {
	selection := &promptui.Select{
		Label: "Please choose the subscription ID you would like to use.",
		Items: subIds,
	}

	_, selectResponse, err := selection.Run()
	if err != nil {
		return err.Error()
	}

	return selectResponse
}

func init() {
	rootCmd.AddCommand(newSetUpCmd())
}
