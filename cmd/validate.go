package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
)

type validateCmd struct {
	manifestPath    string
	imagePullSecret bool
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
	f.BoolVarP(&vc.imagePullSecret, "imagePullSecret", "s", false, "'imagePullSecret' enables the Safeguard that checks for usage of an image pull secret within the manifest(s)")

	return cmd
}

// run is our entry point to GetManifestResults
func (vc *validateCmd) run(c *cobra.Command) error {
	if vc.manifestPath == "" {
		return fmt.Errorf("path to the manifests cannot be empty")
	}

	// AddSafeguardCRIP just adds Container Restricted Image Pulls to the list of safeguards the client will review
	// against the given manifest
	if vc.imagePullSecret {
		safeguards.AddSafeguardCRIP()
	}

	ctx := context.Background()
	isDir, err := safeguards.IsDirectory(vc.manifestPath)
	if err != nil {
		return fmt.Errorf("not a valid file or directory: %w", err)
	}

	var manifestFiles []safeguards.ManifestFile
	if isDir {
		manifestFiles, err = safeguards.GetManifestFiles(vc.manifestPath)
		if err != nil {
			return err
		}
	} else if safeguards.IsYAML(vc.manifestPath) {
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

	manifestViolationsFound := false
	for _, v := range manifestViolations {
		log.Printf("Analyzing %s for violations", v.Name)
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

	if manifestViolationsFound {
		c.SilenceUsage = true
		return fmt.Errorf("violations found")
	} else {
		log.Printf("✅ No violations found in \"%s\".", vc.manifestPath)
	}

	return nil
}
