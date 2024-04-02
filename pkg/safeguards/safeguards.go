package safeguards

import (
	"context"
	"embed"

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

// GetManifestViolations takes in a list of manifest files and returns a slice of ManifestViolation structs
func GetManifestViolations(ctx context.Context, manifestFiles []ManifestFile) ([]ManifestViolation, error) {
	var manifestViolations = make([]ManifestViolation, 0)

	// constraint client instantiation
	c, err := getConstraintClient()
	if err != nil {
		return manifestViolations, err
	}

	// retrieval of templates, constraints, and deployment
	constraintTemplates, err := fc.ReadConstraintTemplates()
	if err != nil {
		return manifestViolations, err
	}
	constraints, err := fc.ReadConstraints()
	if err != nil {
		return manifestViolations, err
	}

	// loading of templates, constraints into constraint client
	err = loadConstraintTemplates(ctx, c, constraintTemplates)
	if err != nil {
		return manifestViolations, err
	}
	err = loadConstraints(ctx, c, constraints)
	if err != nil {
		return manifestViolations, err
	}

	// organized map of manifest object by file name
	manifestMap := make(map[string][]*unstructured.Unstructured, 0)

	// aggregate of every manifest object into one list
	objects := []*unstructured.Unstructured{}

	for _, m := range manifestFiles {
		objs, err := fc.ReadManifests(m.Path) // read all the objects stored in a single file
		if err != nil {
			log.Errorf("reading objects %s", err.Error())
			return manifestViolations, err
		}

		objects = append(objects, objs...)
		manifestMap[m.Name] = objs
	}

	// thbarnes: loadData loads manifest data into client for review as well
	if len(objects) > 0 {
		err = loadData(ctx, c, objects)
	}

	for _, m := range manifestFiles {
		var objectViolations map[string][]string

		// validation of deployment manifest with constraints, templates loaded
		objectViolations, err = getObjectViolations(ctx, c, manifestMap[m.Name])
		if err != nil {
			log.Errorf("validating objects: %s", err.Error())
			return manifestViolations, err
		}
		if len(objectViolations) > 0 {
			manifestViolations = append(manifestViolations, ManifestViolation{
				Name:             m.Name,
				ObjectViolations: objectViolations,
			})
		}
	}

	return manifestViolations, nil
}
