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
	"os/exec"
	"strconv"
	"encoding/json"

	"github.com/spf13/cobra"
	//"github.com/manifoldco/promptui"
)

type SetUpCmd struct {
	appName string
	subscriptionID string
	resourceGroupName string
}

func newConnectCmd() *cobra.Command {
	sc := &SetUpCmd{}

	// setup-ghCmd represents the setup-gh command
	var cmd = &cobra.Command{
		Use:   "setup-gh",
		Short: "automates setting up Github OIDC",
		Long: `This command automates the process of setting up Github OIDC by creating an Azure Active Directory application 
		and service principle, and configuring that application to trust github`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print("")
			sc.ValidateSetUpConfig()
			return sc.CreateServiceProvider()
		},
	}

	f := cmd.Flags()

	f.StringVarP(&sc.appName, "app", "a", "myRandomApp", "name of Azure Active Directory application")
	f.StringVarP(&sc.subscriptionID, "subscription-id", "s", "", "the Azure subscription ID")
	f.StringVarP(&sc.resourceGroupName, "resource-group-name", "r", "myNewResourceGroup", "the name of the Azure resource group")
	cmd.MarkFlagRequired("subscription-id")

	return cmd
}



// func getAppName() string {
// 	validate := func(input string) error {
// 		if input == "" {
// 			return errors.New("Invalid app name")
// 		}
// 		return nil
// 	}

// 	prompt := promptui.Prompt{
// 		Label:    "Enter Azure Active Directory app name",
// 		Validate: validate,
// 	}

// 	result, err := prompt.Run()

// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return err.Error()
// 	}

// 	return result
// }

// func getSubscriptionID() string {
	// validate := func(input string) error {
	// 	_, err := strconv.ParseFloat(input, 64)
	// 	if err != nil {
	// 		return errors.New("Invalid number")
	// 	}
	// 	return nil
	// }

// 	prompt := promptui.Prompt{
// 		Label:    "Number",
// 		Validate: validate,
// 	}

// 	result, err := prompt.Run()

// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return err.Error()
// 	}

// 	return result
// }

// func getResourceGroup() string {
// 	validate := func(input string) error {
// 		if input == "" {
// 			return errors.New("Invalid resource group name")
// 		}
// 		return nil
// 	}

// 	prompt := promptui.Prompt{
// 		Label:    "Enter Azure resource group name",
// 		Validate: validate,
// 	}

// 	result, err := prompt.Run()

// 	if err != nil {
// 		fmt.Printf("Prompt failed %v\n", err)
// 		return err.Error()
// 	}

// 	return result
// }

func (sc *SetUpCmd) setAZContext() error {
	setContextCmd := exec.Command("az", "account", "set", "--subscription", sc.subscriptionID)
	stdoutStderr, err := setContextCmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", stdoutStderr)

	return nil
}


func (sc *SetUpCmd) CreateServiceProvider() error {
	// TODO: set context to correct subscription
	// if err := sc.setAZContext(); err != nil {
	// 	return err
	// }

	// createAppCmd := exec.Command("az", "ad", "app", "create", "--display-name", sc.appName)
	// using the az show app command for testing purposes 
	createAppCmd := exec.Command("az", "ad", "app", "show","--only-show-errors", "--id", "864b58c9-1c86-4e22-a472-f866438378d0")
	stdoutStderr, err := createAppCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", stdoutStderr)
		return err
	}

	var result map[string]interface{}  
    json.Unmarshal(stdoutStderr, &result)
	appId := result["appId"]
	fmt.Println(appId)
	

	return nil
}

func (sc *SetUpCmd) ValidateSetUpConfig() error {
	//fmt.Printf("%v", sc)

	// TODO: check subscriptionID length
	_, err := strconv.ParseFloat(sc.subscriptionID, 64)
		if err != nil {
			return errors.New("Invalid number")
		}
	
	if sc.appName == "" {
		return errors.New("Invalid app name")
	} else if sc.resourceGroupName == "" {
		return errors.New("Invalid resource group name")
	}
	
	return nil
}

func init() {
	rootCmd.AddCommand(newConnectCmd())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setup-ghCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setup-ghCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
