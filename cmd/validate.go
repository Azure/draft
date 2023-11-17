package cmd

import (
	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type validateCmd struct {
	safeguardsOnly bool
	manifestPath   string
}

func newValidateCmd() *cobra.Command {
	vc := &validateCmd{}
	var cmd = &cobra.Command{
		Use:   "validate",
		Short: "Validates deployment manifests against AKS best practices",
		Long:  `This command validates deployment manifests against several AKS best practices.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := vc.run(); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	// thbarnes: add validation to the path
	f.StringVarP(&vc.manifestPath, "manifest", "m", "", "'manifest' asks for the path to the deployment manifest")
	f.BoolVarP(&vc.safeguardsOnly, "safeguards-only", "sg", false, "'safeguards-only' asserts whether or not validate will only run against safeguards constraints")

	return cmd
}

func (vc *validateCmd) run() error {
	log.Debugf("validating deployment manifest")

	err := safeguards.ValidateDeployment(vc.manifestPath, "")
	if err != nil {
		log.Errorf("validating safeguards: %s", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newValidateCmd())
}
