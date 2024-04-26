package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
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
			if err := vc.run(cmd); err != nil {
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

// isYAML determines if a file is of the YAML extension or not
func isYAML(path string) bool {
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}

// getManifests uses filepath.Walk to retrieve a list of the manifest files within the given manifest path
func getManifestFiles(p string) ([]safeguards.ManifestFile, error) {
	var manifestFiles []safeguards.ManifestFile

	noYamlFiles := true
	err := filepath.Walk(p, func(walkPath string, info fs.FileInfo, err error) error {
		manifest := safeguards.ManifestFile{}
		// skip when walkPath is just given path and also a directory
		if p == walkPath && info.IsDir() {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking path %s with error: %w", walkPath, err)
		}

		if !info.IsDir() && info.Name() != "" && isYAML(walkPath) {
			log.Debugf("%s is not a directory, appending to manifestFiles", info.Name())
			noYamlFiles = false

			manifest.Name = info.Name()
			manifest.Path = walkPath
			manifestFiles = append(manifestFiles, manifest)
		} else if !isYAML(p) {
			log.Debugf("%s is not a manifest file, skipping...", info.Name())
		} else {
			log.Debugf("%s is a directory, skipping...", info.Name())
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk directory: %w", err)
	}
	if noYamlFiles {
		return nil, fmt.Errorf("no manifest files found within given path")
	}

	return manifestFiles, nil
}

// run is our entry point to GetManifestResults
func (vc *validateCmd) run(c *cobra.Command) error {
	if vc.manifestPath == "" {
		return fmt.Errorf("path to the manifests cannot be empty")
	}

	ctx := context.Background()
	isDir, err := isDirectory(vc.manifestPath)
	if err != nil {
		return fmt.Errorf("not a valid file or directory: %w", err)
	}

	var manifestFiles []safeguards.ManifestFile
	if isDir {
		manifestFiles, err = getManifestFiles(vc.manifestPath)
		if err != nil {
			return err
		}
	} else if isYAML(vc.manifestPath) {
		manifestFiles = append(manifestFiles, safeguards.ManifestFile{
			Name: path.Base(vc.manifestPath),
			Path: vc.manifestPath,
		})
	} else {
		return fmt.Errorf("expected at least one .yaml or .yml file within given path")
	}

	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	log.Debugf("validating manifests")
	manifestViolations, err := safeguards.GetManifestResults(ctx, manifestFiles)
	if err != nil {
		log.Errorf("validating safeguards: %s", err.Error())
		return err
	}

	for _, v := range manifestViolations {
		log.Printf("Analyzing %s for violations", v.Name)
		manifestViolationsFound := false
		// returning the full list of violations after each manifest is checked
		for file, violations := range v.ObjectViolations {
			log.Printf("  %s:", file)
			for _, violation := range violations {
				log.Printf("    ❌ %s", violation)
				manifestViolationsFound = true
			}
		}
		if !manifestViolationsFound {
			log.Printf("    ✅ no violations found.")
		}
	}

	if len(manifestViolations) > 0 {
		c.SilenceUsage = true
	} else {
		log.Printf("✅ No violations found in \"%s\".", vc.manifestPath)
	}

	return nil
}
