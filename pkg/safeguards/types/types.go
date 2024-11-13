package types

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"

	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	api "github.com/open-policy-agent/gatekeeper/v3/apis"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type FileCrawler struct {
	Safeguards   []Safeguard
	ConstraintFS fs.FS
}

type Safeguard struct {
	Name           string
	TemplatePath   string
	ConstraintPath string
}

type ManifestFile struct {
	Name            string
	ManifestContent []byte
}

type ManifestResult struct {
	Name             string              // the name of the manifest
	ObjectViolations map[string][]string // a map of string object names to slice of string objectViolations
	ViolationsCount  int                 // a count of how many violations are associated with this manifest
}

// methods for retrieval of manifest, constraint templates, and constraints
func (fc FileCrawler) ReadManifests(manifestBytes []byte) ([]*unstructured.Unstructured, error) {
	// Create a new bytes.Reader from the byte slice
	bufReader := bufio.NewReader(bytes.NewReader(manifestBytes))

	// Read the Kubernetes resources using the reader
	manifests, err := reader.ReadK8sResources(bufReader)
	if err != nil {
		return nil, fmt.Errorf("reading manifests: %w", err)
	}

	return manifests, nil
}

func GetScheme() *runtime.Scheme {
	var s = runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	_ = api.AddToScheme(s)
	return s
}

func (fc FileCrawler) ReadConstraintTemplates() ([]*templates.ConstraintTemplate, error) {
	var constraintTemplates []*templates.ConstraintTemplate

	for _, sg := range fc.Safeguards {
		ct, err := reader.ReadTemplate(GetScheme(), fc.ConstraintFS, sg.TemplatePath)
		if err != nil {
			return nil, fmt.Errorf("reading template: %w", err)
		}
		constraintTemplates = append(constraintTemplates, ct)
	}

	return constraintTemplates, nil
}

func (fc FileCrawler) ReadConstraintTemplate(name string) (*templates.ConstraintTemplate, error) {
	var constraintTemplate *templates.ConstraintTemplate

	for _, sg := range fc.Safeguards {
		if sg.Name == name {
			ct, err := reader.ReadTemplate(GetScheme(), fc.ConstraintFS, sg.TemplatePath)
			if err != nil {
				return nil, fmt.Errorf("could not read template: %w", err)
			}
			constraintTemplate = ct
		}
	}
	if constraintTemplate == nil {
		return nil, fmt.Errorf("no constraint template exists with name: %s", name)
	}

	return constraintTemplate, nil
}

func (fc FileCrawler) ReadConstraints() ([]*unstructured.Unstructured, error) {
	var constraints []*unstructured.Unstructured

	for _, sg := range fc.Safeguards {
		u, err := reader.ReadConstraint(fc.ConstraintFS, sg.ConstraintPath)
		if err != nil {
			return nil, fmt.Errorf("could not add constraint: %w", err)
		}

		constraints = append(constraints, u)
	}

	return constraints, nil
}

func (fc FileCrawler) ReadConstraint(name string) (*unstructured.Unstructured, error) {
	var constraint *unstructured.Unstructured

	for _, sg := range fc.Safeguards {
		if sg.Name == name {
			c, err := reader.ReadConstraint(fc.ConstraintFS, sg.ConstraintPath)
			if err != nil {
				return nil, fmt.Errorf("could not add constraint: %w", err)
			}

			constraint = c
		}
	}
	if constraint == nil {
		return nil, fmt.Errorf("no constraint exists with name: %s", name)
	}

	return constraint, nil
}
