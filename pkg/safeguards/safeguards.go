package safeguards

import (
	"context"
	"embed"

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

// ValidateManifests takes in a list of manifest files and returns a map of manifestFiles/dirs to map of manifestNames to violation strings
func ValidateManifests(ctx context.Context, manifestFiles []string) (map[string]map[string][]string, error) {
	var manifestFileViolations = make(map[string]map[string][]string)

	// constraint client instantiation
	c, err := getConstraintClient()
	if err != nil {
		return manifestFileViolations, err
	}

	// retrieval of templates, constraints, and deployment
	constraintTemplates, err := fc.ReadConstraintTemplates()
	if err != nil {
		return manifestFileViolations, err
	}
	constraints, err := fc.ReadConstraints()
	if err != nil {
		return manifestFileViolations, err
	}

	// loading of templates, constraints into constraint client
	err = loadConstraintTemplates(ctx, c, constraintTemplates)
	if err != nil {
		return manifestFileViolations, err
	}
	err = loadConstraints(ctx, c, constraints)
	if err != nil {
		return manifestFileViolations, err
	}

	for _, m := range manifestFiles {
		var fileViolations map[string][]string
		manifests, err := fc.ReadManifests(m) // read all the manifests stored in a single file
		if err != nil {
			log.Errorf("reading manifests %s", err.Error())
			return manifestFileViolations, err
		}

		// validation of deployment manifest with constraints, templates loaded
		fileViolations, err = validateManifests(ctx, c, manifests)
		if err != nil {
			log.Errorf("validating manifests: %s", err.Error())
			return manifestFileViolations, err
		}
		manifestFileViolations[m] = fileViolations
	}

	return manifestFileViolations, nil
}
