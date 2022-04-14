package cmd

import (
	"errors"
	"fmt"
	"strings"

	
	"github.com/Azure/draftv2/pkg/providers"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
)


func newSetUpCmd() *cobra.Command {
	sc := &providers.SetUpCmd{}

	// setup-ghCmd represents the setup-gh command
	var cmd = &cobra.Command{
		Use:   "setup-gh",
		Short: "Automates the Github OIDC setup process",
		Long: `This command will automate the Github OIDC setup process by creating an Azure Active Directory 
application and service principle, and will configure that application to trust github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fillSetUpConfig(sc)

			log.Info("--> Setting up Github OIDC...")
			
			if err := runProviderSetUp(sc); err != nil {
				return err
			}

			log.Info("Draft has successfully set up Github OIDC for your project 😃")
			log.Into("Use 'draft generate-workflow' to generate a Github workflow to build and deploy an application on AKS.")

			return nil		
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.AppName, "app", "a", "", "Specify the name of the Azure Active Directory application")
	f.StringVarP(&sc.SubscriptionID, "subscription-id", "s", "", "Specify the Azure subscription ID")
	f.StringVarP(&sc.ResourceGroupName, "resource-group-name", "r", "", "Specify the name of the Azure resource group")
	f.StringVarP(&sc.Provider, "provider", "p", "", "Specify the cloud provider")
	f.StringVarP(&sc.Repo, "gh-repo", "g", "", "Specify the github repository link")

	return cmd
}

func fillSetUpConfig(sc *providers.SetUpCmd) {
	if sc.Provider == "" {
		sc.Provider = getCloudProvider()
	}

	if sc.AppName == "" {
		sc.AppName = getAppName()
	}

	if sc.SubscriptionID == "" {
		sc.SubscriptionID = getSubscriptionID()
	}

	if sc.ResourceGroupName == "" {
		sc.ResourceGroupName = getResourceGroup()
	}

	if sc.Repo == "" {
		sc.Repo = getGhRepo()
	}
}


func runProviderSetUp(sc *providers.SetUpCmd) error {
	provider := strings.ToLower(sc.Provider)
	if provider == "azure" {
		// call azure provider logic
		return providers.InitiateAzureOIDCFlow(sc)
			
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
		Label:    "Enter app name",
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


func init() {
	rootCmd.AddCommand(newSetUpCmd())

}
