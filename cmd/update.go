package cmd

import (
	"errors"

	"github.com/Azure/draft/pkg/web"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	sa := &web.ServiceAnnotations{}

	// updateCmd represents the update command
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Updates your application to be internet accessible",
		Long: `This command automatically updates your yaml files as necessary so that your application
		will be able to receive external requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fillUpdateConfig(sa)

			if err := web.UpdateServiceFile(sa, "."); err != nil {
				return err
			}
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&sa.Host, "host", "a", "", "the host of the ingress resource")
	f.StringVarP(&sa.Cert, "cert", "s", "", "the URI of the Keyvault certificate to present")

	return cmd

}

func fillUpdateConfig(sa *web.ServiceAnnotations) {
	if sa.Host == "" {
		sa.Host = getHost()
	}

	if sa.Cert == "" {
		sa.Cert = getCert()
	}
}

func getHost() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("Invalid host")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter ingress resource host",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return err.Error()
	}

	return result
}

func getCert() string {
	validate := func(input string) error {
		if input == "" {
			return errors.New("Invalid cert")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Enter URI of the Keyvault certificate",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		return err.Error()
	}

	return result
}

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
