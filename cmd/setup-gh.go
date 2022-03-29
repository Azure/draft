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

	"github.com/Azure/draftv2/pkg/providers"
	"github.com/spf13/cobra"
)



func newConnectCmd() *cobra.Command {
	sc := &providers.SetUpCmd{}

	// setup-ghCmd represents the setup-gh command
	var cmd = &cobra.Command{
		Use:   "setup-gh",
		Short: "automates setting up Github OIDC",
		Long: `This command automates the process of setting up Github OIDC by creating an Azure Active Directory application 
		and service principle, and configuring that application to trust github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			
			if err := hasValidProviderInfo(sc); err != nil {
				return err
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
	f.StringVarP(&sc.Provider, "provider", "p", "", "your cloud provider")

	return cmd
}


func hasValidProviderInfo(sc *providers.SetUpCmd) error {
	if sc.Provider == "azure" && sc.SubscriptionID == "" {
		return errors.New("If provider is azure, must provide azure subscription ID")
	} 

	return nil
}



func runProviderSetUp(sc *providers.SetUpCmd) error {
	if sc.Provider == "azure" {
		// call azure provider logic
		if err := providers.InitiateAzureOIDCFlow(sc); err != nil {
			return err
		}
	} else {
		// call logic for user-submitted provider
		fmt.Printf("The provider is %v\n", sc.Provider)
	}

	return nil
}


func init() {
	rootCmd.AddCommand(newConnectCmd())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//setup-ghCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setup-ghCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
