package cmd

import (
	"bufio"
	"fmt"
	"github.com/Azure/draft/pkg/providers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func newCleanupCmd() *cobra.Command {
	sc := &providers.SetUpCmd{}

	var cmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Cleans up Azure resources created during the setup process",
		Long: `This command will clean up the Azure AD application and its associated resources.
Ensure that the application and resources are not in use before proceeding.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cleanup(sc)
			if err != nil {
				return err
			}

			log.Info("the App was deleted successfully")

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sc.AppName, "app", "a", emptyDefaultFlagValue, "specify the Azure Active Directory application name")
	cmd.MarkFlagRequired("app")
	sc.Provider = provider

	return cmd
}

func cleanup(sc *providers.SetUpCmd) error {
	fmt.Println("WARNING: This operation will delete the Azure AD application and its associated resources.")
	fmt.Println("Ensure that you are not deleting an application in use.")
	fmt.Print("Do you want to proceed? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if strings.ToLower(response) == "yes" {
		err := sc.CleanUpAzureResources(sc.AppName)
		if err != nil {
			log.Fatalf("Failed to clean up resources: %v", err)
		} else {
			log.Info("Successfully cleaned up Azure resources.")
		}
	} else {
		fmt.Println("Cleanup operation aborted.")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(newCleanupCmd())
}
