package cmd

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/draftv2/pkg/osutil"
	"github.com/Azure/draftv2/pkg/providers"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)


func newSetUpCmd() *cobra.Command {
	sc := &providers.SetUpCmd{}

	// setup-ghCmd represents the setup-gh command
	var cmd = &cobra.Command{
		Use:   "setup-gh",
		Short: "automates setting up Github OIDC",
		Long: `This command automates the process of setting up Github OIDC by creating an Azure Active Directory application 
		and service principle, and configuring that application to trust github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagsAreSet(cmd.Flags()) {
				gatherUserInfo(sc)
			} else {
				if err := hasValidProviderInfo(sc); err != nil {
					return err
				}
			}
			
			if err := runProviderSetUp(sc); err != nil {
				return err
			}

			return nil		
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.AppName, "app", "a", "myRandomApp", "name of Azure Active Directory application")
	f.StringVarP(&sc.SubscriptionID, "subscription-id", "s", "", "the Azure subscription ID")
	f.StringVarP(&sc.ResourceGroupName, "resource-group-name", "r", "myNewResourceGroup", "the name of the Azure resource group")
	f.StringVarP(&sc.Provider, "provider", "p", "azure", "your cloud provider")
	f.StringVarP(&sc.Repo, "gh-repo", "g", "", "your github repo")

	return cmd
}


func hasValidProviderInfo(sc *providers.SetUpCmd) error {
	// TODO: move validate set up config here?
	if sc.Repo == "" {
		return errors.New("Must provide github repo")
	}

	provider := strings.ToLower(sc.Provider)
	if provider == "azure" {
		osutil.CheckAzCliInstalled()
		if !osutil.IsLoggedInToAz() {
			log.Fatal("Error: Must be logged in to az cli. Run the az --help command for more information on logging in via cli")
		}

		if sc.SubscriptionID == "" {
			return errors.New("If provider is azure, must provide azure subscription ID")
		}
	} 

	return nil
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

func flagsAreSet(f *pflag.FlagSet) bool {
	return f.Changed("gh-repo") || f.Changed("subscription-id") || f.Changed("provider")
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
		Label:    "Enter github organization and repo; example: organization/repoName",
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

func gatherUserInfo(sc *providers.SetUpCmd) {
	if getCloudProvider() == "azure" {
		osutil.CheckAzCliInstalled()
		if !osutil.IsLoggedInToAz() {
			log.Fatal("Error: Must be logged in to az cli. Run the az --help command for more information on logging in via cli")
		}

		sc.AppName = getAppName()
		sc.SubscriptionID = getSubscriptionID()
		sc.ResourceGroupName = getResourceGroup()
		sc.Repo = getGhRepo()
	} else {
		// prompts for other cloud providers
	}

}


func init() {
	rootCmd.AddCommand(newSetUpCmd())

}
