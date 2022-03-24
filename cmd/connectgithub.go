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
	"fmt"

	"github.com/spf13/cobra"
)


func newConnectCmd() *cobra.Command {
	// connectgithubCmd represents the connectgithub command
	var cmd = &cobra.Command{
		Use:   "connectgithub",
		Short: "automates setting up Github OIDC",
		Long: `This command automates the process of setting up Github OIDC by creating an Azure Active Directory application 
		and service principle, and configuring that application to trust github`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("")
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(newConnectCmd())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectgithubCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectgithubCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
