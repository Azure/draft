package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

	f.StringVarP(&vc.manifestPath, "manifest", "m", "", "'manifest' asks for the path to the manifest")

	return cmd
}

// return file pointer
// use os.stat to validate path instead of regexp
//func validatePath(path string) (bool, error) {
//	isValidPath, _ := regexp.MatchString("^(.+)/([^/]+)$", path)
//	if !isValidPath {
//		return false, fmt.Errorf("'%s' is not a valid path", path)
//	}
//
//	return true, nil
//}

// isDirectory determines if a file represented by path is a directory or not
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func (vc *validateCmd) run() error {
	ctx := context.Background()

	isDir, err := isDirectory(vc.manifestPath)
	if err != nil {
		return fmt.Errorf("could not determine if given path is a directory: %w", err)
	}

	var manifests []string
	// use fs.WalkDir
	if isDir {
		// -> append to manifests
		err = filepath.Walk(vc.manifestPath, func(path string, info fs.FileInfo, err error) error {
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
			return fmt.Errorf("could not walk directory: %w", err)
		}
	} else {
		manifests = append(manifests, vc.manifestPath)
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
