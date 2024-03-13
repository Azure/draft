package safeguards

import (
	"bufio"
	"context"
	"fmt"
	"os"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/rego"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/target"
	log "github.com/sirupsen/logrus"

	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// retrieves the constraint client that does all rego code related operations
func getConstraintClient() (*constraintclient.Client, error) {
	driver, err := rego.New()
	if err != nil {
		return nil, fmt.Errorf("could not create rego driver: %w", err)
	}

	c, err := constraintclient.NewClient(constraintclient.Targets(&target.K8sValidationTarget{}), constraintclient.Driver(driver))
	if err != nil {
		return nil, fmt.Errorf("could not create constraint client: %w", err)
	}

	return c, nil
}

// sorts the list of supported safeguards versions and returns the last item in the list
func getLatestSafeguardsVersion() string {
	semver.Sort(supportedVersions)
	return supportedVersions[len(supportedVersions)-1]
}

func updateSafeguardPaths() {
	for _, sg := range safeguards {
		sg.templatePath = fmt.Sprintf("%s/%s/%s", selectedVersion, sg.name, templateFileName)
		sg.constraintPath = fmt.Sprintf("%s/%s/%s", selectedVersion, sg.name, constraintFileName)
	}
}

// methods for retrieval of manifest, constraint templates, and constraints
func (fc FileCrawler) ReadManifests(path string) ([]*unstructured.Unstructured, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %q: %w", path, err)
	}
	defer file.Close()

	manifests, err := reader.ReadK8sResources(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", path, err)
	}

	return manifests, nil
}

func (fc FileCrawler) ReadConstraintTemplates() ([]*templates.ConstraintTemplate, error) {
	var constraintTemplates []*templates.ConstraintTemplate

	for _, sg := range fc.Safeguards {
		ct, err := reader.ReadTemplate(s, fc.constraintFS, sg.templatePath)
		if err != nil {
			return nil, fmt.Errorf("could not read template: %w\n", err)
		}
		constraintTemplates = append(constraintTemplates, ct)
	}

	return constraintTemplates, nil
}

func (fc FileCrawler) ReadConstraintTemplate(name string) (*templates.ConstraintTemplate, error) {
	var constraintTemplate *templates.ConstraintTemplate

	for _, sg := range fc.Safeguards {
		if sg.name == name {
			ct, err := reader.ReadTemplate(s, fc.constraintFS, sg.templatePath)
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
		u, err := reader.ReadConstraint(fc.constraintFS, sg.constraintPath)
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
		if sg.name == name {
			c, err := reader.ReadConstraint(fc.constraintFS, sg.constraintPath)
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

// loads constraint templates, constraints into constraint client
func loadConstraintTemplates(ctx context.Context, c *constraintclient.Client, constraintTemplates []*templates.ConstraintTemplate) error {
	// AddTemplate adds the template source code to OPA and registers the CRD with the client for
	// schema validation on calls to AddConstraint. On error, the responses return value
	// will still be populated so that partial results can be analyzed.
	for _, ct := range constraintTemplates {
		_, err := c.AddTemplate(ctx, ct)
		if err != nil {
			return fmt.Errorf("could not add template: %w", err)
		}
	}

	return nil
}

func loadConstraints(ctx context.Context, c *constraintclient.Client, constraints []*unstructured.Unstructured) error {
	// AddConstraint validates the constraint and, if valid, inserts it into OPA.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.
	for _, con := range constraints {
		_, err := c.AddConstraint(ctx, con)
		if err != nil {
			return fmt.Errorf("could not add constraint: %w", err)
		}
	}

	return nil
}

// does validation on manifest based on loaded constraint templates, constraints
func validateManifests(ctx context.Context, c *constraintclient.Client, manifests []*unstructured.Unstructured) ([]string, error) {
	// Review makes sure the provided object satisfies all stored constraints.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.
	var violations []string
	for _, manifest := range manifests {
		log.Printf("Reviewing %s...", manifest.GetName())
		res, err := c.Review(ctx, manifest)
		if err != nil {
			return violations, fmt.Errorf("could not review manifests: %w", err)
		}

		for _, v := range res.ByTarget {
			log.Printf("Found %d errors", len(v.Results))
			for _, result := range v.Results {
				if result.Msg != "" {
					log.Printf("violation: %s", result.Msg)
					violations = append(violations, result.Msg)
				}
			}
		}
		fmt.Println("\n")
	}

	return violations, nil
}
