package safeguards

import (
	"context"
	"embed"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/safeguards/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//go:embed lib
var embedFS embed.FS

var fc types.FileCrawler

// primes the scheme to be able to interpret beta templates
func init() {

	types.SelectedVersion = getLatestSafeguardsVersion()
	updateSafeguardPaths(&types.Safeguards)

	fc = types.FileCrawler{
		Safeguards:   types.Safeguards,
		ConstraintFS: embedFS,
	}
}

type ManifestResult struct {
	Name             string              // the name of the manifest
	ObjectViolations map[string][]string // a map of string object names to slice of string objectViolations
	ViolationsCount  int                 // a count of how many violations are associated with this manifest
}

// GetManifestResults takes in a list of manifest files and returns a slice of ManifestViolation structs
func GetManifestResults(ctx context.Context, manifestFiles []types.ManifestFile) ([]types.ManifestResult, error) {
	if len(manifestFiles) == 0 {
		return nil, fmt.Errorf("path cannot be empty")
	}

	manifestResults := make([]types.ManifestResult, 0)

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
		manifestObjects, err := fc.ReadManifests(m.ManifestContent) // read all the objects stored in a single file
		if err != nil {
			log.Errorf("reading objects %s", err.Error())
			return manifestResults, err
		}

		allManifestObjects = append(allManifestObjects, manifestObjects...)
		manifestMap[m.Name] = manifestObjects
	}

	if len(allManifestObjects) > 0 {
		err := loadManifestObjects(ctx, c, allManifestObjects)
		if err != nil {
			return manifestResults, err
		}
	}

	for _, m := range manifestFiles {
		var objectViolations map[string][]string

		// validation of deployment manifest with constraints, templates loaded
		objectViolations, err = getObjectViolations(ctx, c, manifestMap[m.Name])
		if err != nil {
			log.Errorf("validating objects: %s", err.Error())
			return manifestResults, err
		}

		manifestResults = append(manifestResults, types.ManifestResult{
			Name:             m.Name,
			ObjectViolations: objectViolations,
			ViolationsCount:  len(objectViolations),
		})
	}

	return manifestResults, nil
}
