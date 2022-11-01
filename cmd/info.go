package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Format string

const (
	JSON Format = "json"
)

type infoCmd struct {
	format string
}

func newInfoCmd() *cobra.Command {
	ic := &infoCmd{}
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Prints draft supported values in machine-readable format",
		Long:  `This command prints information about the current draft environment and supported values such as supported dockerfile languages and deployment manifest types.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ic.run(); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&ic.format, "format", "f", ".", "specify the format to print draft information in (json, yaml, etc)")

	return cmd
}

func (uc *infoCmd) run() error {
	log.Println("infoCmd.run() called with format: ", uc.format)
	return nil
}

func init() {
	rootCmd.AddCommand(newInfoCmd())
}
