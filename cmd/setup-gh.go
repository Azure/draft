package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/draft/pkg/providers"
	"github.com/Azure/draft/pkg/spinner"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			fillSetUpConfig(sc)

			s := spinner.GetSpinner("--> Setting up Github OIDC...")
			s.Start()
			err := runProviderSetUp(sc)
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
	f.StringVarP(&sc.AppName, "app", "a", "", "specify the Azure Active Directory application name")
	f.StringVarP(&sc.SubscriptionID, "subscription-id", "s", "", "specify the Azure subscription ID")
	f.StringVarP(&sc.ResourceGroupName, "resource-group", "r", "", "specify the Azure resource group name")
	f.StringVarP(&sc.Repo, "gh-repo", "g", "", "specify the github repository link")
	sc.Provider = provider
	return cmd
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
