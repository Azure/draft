package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"regexp"
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

	f.StringVarP(&vc.manifestPath, "manifest", "m", "", "'manifest' asks for the path to the manifest")

	return cmd
}

// return file pointer
// use os.stat to validate path instead of regexp
func validatePath(path string) (bool, error) {
	isValidPath, _ := regexp.MatchString("^(.+)/([^/]+)$", path)
	if !isValidPath {
		return false, fmt.Errorf("'%s' is not a valid path", path)
	}

	return true, nil
}

func (vc *validateCmd) run() error {
	ctx := context.Background()

	log.Debugf("validating given path")
	isDir, err := validatePath(vc.manifestPath)
	if err != nil {
		log.Errorf("validating path: %s", err)
		return err
	}

	var manifests []string
	// use os.walk
	if isDir {
		// -> append to manifests
	} else {
		// -> append one file to manifests
	}

	// use manifests here instead, update name
	log.Debugf("validating manifest")
	err = safeguards.ValidateManifest(ctx, vc.manifestPath)
	if err != nil {
		log.Errorf("validating safeguards: %s", err)
		return err
	}

	return nil
}
