package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/addons"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

func newUpdateCmd() *cobra.Command {
	dest := ""
	provider := ""
	addon := ""
	userInputs := make(map[string]string)
	templateWriter := &writers.LocalFSWriter{}
	// updateCmd represents the update command
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Updates your application to be internet accessible",
		Long: `This command automatically updates your yaml files as necessary so that your application
will be able to receive external requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := addons.GenerateAddon(template.Addons, provider, addon, dest, userInputs, templateWriter); err != nil {
				return err
			}

			log.Info("Draft has successfully updated your yaml files so that your application will be able to receive external requests 😃")

			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringVarP(&provider, "provider", "p", "azure", "cloud provider")
	f.StringVarP(&addon, "addon", "a", "", "addon name")
	return cmd

}

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
