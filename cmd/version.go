package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

const version = "v0.0.6"

func newVersionCmd() *cobra.Command {
	// updateCmd represents the update command
	var updateCmd = &cobra.Command{
		Use:   "version",
		Short: "Get current version of Draft",
		Long:  `Returns the running version of Draft`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("verison: %s", version)
			return nil
		},
	}

	return updateCmd

}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
