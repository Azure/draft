package safeguards

import (
	"context"
	"embed"
	"fmt"

	api "github.com/open-policy-agent/gatekeeper/v3/apis"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

// GetManifestResults takes in a list of manifest files and returns a slice of ManifestViolation structs
func GetManifestResults(ctx context.Context, manifestFiles []ManifestFile) ([]ManifestResult, error) {
	if len(manifestFiles) == 0 {
		return nil, fmt.Errorf("path cannot be empty")
	}

	manifestResults := make([]ManifestResult, 0)

	// constraint client instantiation
	c, err := getConstraintClient()
	if err != nil {
		return manifestResults, err
	}

	// retrieval of templates, constraints, and deployment
	constraintTemplates, err := fc.ReadConstraintTemplates()
	if err != nil {
		return manifestResults, err
	}
	constraints, err := fc.ReadConstraints()
	if err != nil {
		return manifestResults, err
	}

	// loading of templates, constraints into constraint client
	err = loadConstraintTemplates(ctx, c, constraintTemplates)
	if err != nil {
		return manifestResults, err
	}
	err = loadConstraints(ctx, c, constraints)
	if err != nil {
		return manifestResults, err
	}

	// organized map of manifest object by file name
	manifestMap := make(map[string][]*unstructured.Unstructured, 0)
	// aggregate of every manifest object into one list
	allManifestObjects := []*unstructured.Unstructured{}
	for _, m := range manifestFiles {
		manifestObjects, err := fc.ReadManifests(m.Path) // read all the objects stored in a single file
		if err != nil {
			log.Errorf("reading objects %s", err.Error())
			return manifestResults, err
		}

		allManifestObjects = append(allManifestObjects, manifestObjects...)
		manifestMap[m.Name] = manifestObjects
	}

	if len(allManifestObjects) > 0 {
		err = loadManifestObjects(ctx, c, allManifestObjects)
	}

	for _, m := range manifestFiles {
		var objectViolations map[string][]string

		// validation of deployment manifest with constraints, templates loaded
		objectViolations, err = getObjectResults(ctx, c, manifestMap[m.Name])
		if err != nil {
			log.Errorf("validating objects: %s", err.Error())
			return manifestResults, err
		}
		manifestResults = append(manifestResults, ManifestResult{
			Name:             m.Name,
			ObjectViolations: objectViolations,
		})
	}

	return manifestResults, nil
}
