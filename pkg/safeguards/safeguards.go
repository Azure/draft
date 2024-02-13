package safeguards

import (
	"context"
	"os"

	api "github.com/open-policy-agent/gatekeeper/v3/apis"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// Globals
var s = runtime.NewScheme()
var wd, _ = os.Getwd()
var f = os.DirFS(wd)
var fc FileCrawler

// primes the scheme to be able to interpret beta templates
func init() {
	_ = clientgoscheme.AddToScheme(s)
	_ = api.AddToScheme(s)

	selectedVersion = getLatestSafeguardsVersion()
	updateSafeguardPaths()

	fc = FileCrawler{
		Safeguards: safeguards,
	}
}

// ValidatemManifest is what will be called by `draft validate` to validate the user's manifest
// against each safeguards constraint
func ValidateManifest(ctx context.Context, manifestPath string) error {
	// constraint client instantiation
	c, err := getConstraintClient()
	if err != nil {
		return err
	}

	// retrieval of templates, constraints, and deployment
	constraintTemplates, err := fc.ReadConstraintTemplates()
	if err != nil {
		return err
	}
	constraints, err := fc.ReadConstraints()
	if err != nil {
		return err
	}
	manifest, err := fc.ReadManifest(manifestPath)
	if err != nil {
		return err
	}

	// loading of templates, constraints into constraint client
	err = loadConstraintTemplates(ctx, c, constraintTemplates)
	if err != nil {
		return err
	}
	err = loadConstraints(ctx, c, constraints)
	if err != nil {
		return err
	}

	// validation of deployment manifest with constraints, templates loaded
	return validateManifest(ctx, c, manifest)
}
