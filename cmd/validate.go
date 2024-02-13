package cmd

import (
	"context"

	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type validateCmd struct {
	safeguardsOnly bool
	manifestPath   string
}

func init() {
	rootCmd.AddCommand(newValidateCmd())
}

func newValidateCmd() *cobra.Command {
	vc := &validateCmd{}

	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates manifests against AKS best practices",
		Long:  `This command validates manifests against several AKS best practices.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := vc.run(); err != nil {
				return err
			}
			return nil
		},
	}

	f := cmd.Flags()

	// TODO: add validation to the path
	f.StringVarP(&vc.manifestPath, "manifest", "m", "", "'manifest' asks for the path to the manifest")
	f.BoolVarP(&vc.safeguardsOnly, "safeguards-only", "s", false, "'safeguards-only' asserts whether or not validate will only run against safeguards constraints")

	return cmd
}

func (vc *validateCmd) run() error {
	ctx := context.Background()

	log.Debugf("validating manifest")
	err := safeguards.ValidateManifest(ctx, vc.manifestPath)
	if err != nil {
		log.Errorf("validating safeguards: %s", err)
	}

	return nil
}
