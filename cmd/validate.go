package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
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

// getManifests uses fs.WalkDir to retrieve a list of the manifest files within the given manifest path
func getManifestFiles(f fs.FS, p string) ([]string, error) {
	var manifestFiles []string

	err := fs.WalkDir(f, p, func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s with error: %w", filepath, err)
		}

		if !d.IsDir() && d.Name() != "" {
			log.Debugf("%s is not a directory, appending to manifestFiles", d.Name())

			manifestFiles = append(manifestFiles, filepath)
		} else {
			log.Debugf("%s is a directory, skipping...", d.Name())
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk directory: %w", err)
	}

	return manifestFiles, nil
}

// run is our entry point to ValidateManifests
func (vc *validateCmd) run() error {
	if vc.manifestPath == "" {
		return fmt.Errorf("path to the manifests cannot be empty")
	}

	ctx := context.Background()
	isDir, err := isDirectory(vc.manifestPath)
	if err != nil {
		return fmt.Errorf("not a valid file or directory: %w", err)
	}

	var manifestFiles []string
	var manifestFS = os.DirFS(vc.manifestPath)
	if isDir {
		manifestFiles, err = getManifestFiles(manifestFS, vc.manifestPath)
		if err != nil {
			return err
		}
	} else {
		manifestFiles = append(manifestFiles, vc.manifestPath)
	}

	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	log.Debugf("validating manifests")
	err = safeguards.ValidateManifests(ctx, manifestFS, manifestFiles)
	if err != nil {
		log.Errorf("validating safeguards: %s", err.Error())
		return err
	}

	return nil
}
