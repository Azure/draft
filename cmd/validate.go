package cmd

import (
	"github.com/Azure/draft/pkg/guardrails"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type validateCmd struct {
	guardrailsOnly bool
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
	f.BoolVarP(&vc.guardrailsOnly, "guardrails-only", "g", false, "guardrails-only asserts whether or not validate will only run against guardrails constraints")

	return cmd
}

func (vc *validateCmd) run() error {
	log.Debugf("validating deployment manifest")

	err := guardrails.ValidateGuardrailsConstraint()
	if err != nil {
		log.Errorf("validating guardrails: %s", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newValidateCmd())
}
