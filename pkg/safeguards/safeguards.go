package safeguards

import (
	"context"
	"embed"
	"fmt"

	api "github.com/open-policy-agent/gatekeeper/v3/apis"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// Globals
var s = runtime.NewScheme()

//go:embed lib
var embedFS embed.FS

var fc FileCrawler

// primes the scheme to be able to interpret beta templates
func init() {
	_ = clientgoscheme.AddToScheme(s)
	_ = api.AddToScheme(s)

	selectedVersion = getLatestSafeguardsVersion()
	updateSafeguardPaths()

	fc = FileCrawler{
		Safeguards:   safeguards,
		constraintFS: embedFS,
	}
}

// ValidateManifests is what will be called by `draft validate` to validate the user's manifests
// against each safeguards constraint
func ValidateManifests(ctx context.Context, manifestFiles []string) error {
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

	// loading of templates, constraints into constraint client
	err = loadConstraintTemplates(ctx, c, constraintTemplates)
	if err != nil {
		return err
	}
	err = loadConstraints(ctx, c, constraints)
	if err != nil {
		return err
	}

	var violations []string
	for _, m := range manifestFiles {
		manifests, err := fc.ReadManifests(m)
		if err != nil {
			return err
		}

		// validation of deployment manifest with constraints, templates loaded
		err = validateManifests(ctx, c, manifests)
		if err != nil {
			violations = append(violations, err.Error())
		}
	}

	// returning the full list of violations after each manifest is checked
	if len(violations) > 0 {

		return fmt.Errorf("violations have occurred: %s", violations)
	}

	log.Printf("No violations found.")
	return nil
}
