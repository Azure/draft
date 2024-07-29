package cmd

import (
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/Azure/draft/pkg/safeguards"
	"github.com/Azure/draft/pkg/safeguards/preprocessing"
	"github.com/Azure/draft/pkg/safeguards/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type validateCmd struct {
	manifestPath     string
	imagePullSecret  bool
	releaseName      string
	releaseNamespace string
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
	f.StringVarP(&vc.releaseName, "releaseName", "rn", "", "'releaseName' asks for a user-defined release name for the Helm package to use when rendering Helm projects in Draft")
	f.StringVarP(&vc.releaseNamespace, "releaseNamespace", "rns", "", "'releaseNamespace' asks for a user-defined release namespace for the Helm package to use when rendering Helm projects in Draft")

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

	var opt chartutil.ReleaseOptions
	if vc.releaseName != "" {
		opt.Name = vc.releaseName
	}
	if vc.releaseNamespace != "" {
		opt.Namespace = vc.releaseNamespace
	}
	ctx := context.Background()

	var manifestFiles []types.ManifestFile
	manifestFiles, err := preprocessing.GetManifestFiles(vc.manifestPath, opt)
	if err != nil {
		return fmt.Errorf("error retrieving manifest files: %w", err)
	}

	log.Debugf("validating manifests")
	manifestViolations, err := safeguards.GetManifestResults(ctx, manifestFiles)
	if err != nil {
		log.Errorf("validating safeguards: %s", err.Error())
		return err
	}

	anyViolationsFound := false
	for _, v := range manifestViolations {
		log.Printf("Analyzing %s for violations", v.Name)
		manifestHasViolations := false
		// returning the full list of violations after each manifest is checked
		for file, violations := range v.ObjectViolations {
			log.Printf("  %s:", file)
			for _, violation := range violations {
				log.Printf("    ❌ %s", violation)
				anyViolationsFound = true
				manifestHasViolations = true
			}
		}
		if !manifestHasViolations {
			log.Printf("    ✅ no violations found.")
		}
	}

	if anyViolationsFound {
		c.SilenceUsage = true // suppress default Cobra behaviour of printing usage on all errors
		return fmt.Errorf("violations found")
	} else {
		log.Printf("✅ No violations found in \"%s\".", vc.manifestPath)
	}

	return nil
}
