/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/draftv2/pkg/providers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/manifoldco/promptui"
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
			// TODO: check if flags are set - if not, run prompts
			if !flagsAreSet(cmd.Flags()) {
				gatherUserInfo(sc)
				fmt.Printf("%v", sc)
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

	return cmd
}


func hasValidProviderInfo(sc *providers.SetUpCmd) error {
	provider := strings.ToLower(sc.Provider)
	if provider == "azure" && sc.SubscriptionID == "" {
		return errors.New("If provider is azure, must provide azure subscription ID")
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
	return f.Changed("subscription-id")
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
		fmt.Printf("Prompt failed %v\n", err)
		return err.Error()
	}

	return result
}

func getSubscriptionID() string {
	validate := func(input string) error {
		// TODO: check if it's an existing subscription id
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter subscription ID",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
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
		fmt.Printf("Prompt failed %v\n", err)
		return err.Error()
	}

	return result
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
		sc.AppName = getAppName()
		sc.SubscriptionID = getSubscriptionID()
		sc.ResourceGroupName = getResourceGroup()
	} else {
		// prompts for other cloud providers
	}

}


func init() {
	rootCmd.AddCommand(newSetUpCmd())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//setup-ghCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setup-ghCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
