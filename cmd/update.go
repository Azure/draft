
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	// updateCmd represents the update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Updates your application to be internet accessible",
		Long: `This command automatically updates your yaml files as necessary so that your application
		will be able to receive external requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("update called")
			return nil
		},
	}

	return updateCmd

}


func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
