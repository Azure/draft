package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
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

// isDirectory determines if a file represented by path is a directory or not
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {

		return false, err
	}

	return fileInfo.IsDir(), nil
}

// getManifests uses filepath.Walk to retrieve a list of the manifest files within the given manifest path
func getManifests(path string) ([]string, error) {
	var manifests []string

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {

			return fmt.Errorf("error walking path %s with error: %w", path, err)
		}

		if !info.IsDir() {
			log.Debugf("%s is not a directory, appending to manifests", path)
			manifests = append(manifests, path)
		} else {
			log.Debugf("%s is a directory, skipping...", path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk directory: %w", err)
	}

	return manifests, nil
}

// run is our entry point to ValidateManifests
func (vc *validateCmd) run() error {
	if vc.manifestPath == "" {
		return fmt.Errorf("path to the manifests cannot be empty")
	}
	ctx := context.Background()

	isDir, err := isDirectory(vc.manifestPath)
	if err != nil {
		return fmt.Errorf("could not determine if given path is a directory: %w", err)
	}

	var manifests []string
	if isDir {
		manifests, err = getManifests(vc.manifestPath)
		if err != nil {
			return err
		}
	} else {
		manifests = append(manifests, vc.manifestPath)
	}

	log.Debugf("validating manifests")
	err = safeguards.ValidateManifests(ctx, manifests)
	if err != nil {
		log.Errorf("validating safeguards: %s", err)
		return err
	}

	return nil
}
