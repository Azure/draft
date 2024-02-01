package safeguards

import (
	"context"
	"fmt"
	"os"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/rego"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	api "github.com/open-policy-agent/gatekeeper/v3/apis"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/target"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	Constraint_CAI = "container-allowed-images"
	Constraint_CEP = "container-enforce-probes"
	Constraint_CRL = "container-resource-limits"
	Constraint_NUP = "no-unauthenticated-pulls"
	Constraint_PDB = "pod-disruption-budgets"
	Constraint_PEA = "pod-enforce-antiaffinity"
	Constraint_RT  = "restricted-taints"
	Constraint_USS = "unique-service-selectors"
)

type Safeguard struct {
	name           string
	templatePath   string
	constraintPath string
}

type FileCrawler struct{}

var s = runtime.NewScheme()
var wd, _ = os.Getwd()
var f = os.DirFS(wd)

var safeguards = []Safeguard{
	{
		name:           Constraint_CAI,
		templatePath:   "constraints/ContainerAllowedImages/template/container-allowed-images.yaml",
		constraintPath: "constraints/ContainerAllowedImages/constraint/constraint.yaml",
	},
	{
		name:           Constraint_CEP,
		templatePath:   "constraints/ContainerEnforceProbes/template/container-enforce-probes.yaml",
		constraintPath: "constraints/ContainerEnforceProbes/constraint/constraint.yaml",
	},
	{
		name:           Constraint_CRL,
		templatePath:   "constraints/ContainerResourceLimits/template/container-resource-limits.yaml",
		constraintPath: "constraints/ContainerResourceLimits/constraint/constraint.yaml",
	},
	{
		name:           Constraint_NUP,
		templatePath:   "constraints/NoUnauthenticatedPulls/template/no-unauthenticated-pulls.yaml",
		constraintPath: "constraints/NoUnauthenticatedPulls/constraint/constraint.yaml",
	},
	{
		name:           Constraint_PDB,
		templatePath:   "constraints/PodDisruptionBudgets/template/pod-disruption-budgets.yaml",
		constraintPath: "constraints/PodDisruptionBudgets/constraint/constraint.yaml",
	},
	{
		name:           Constraint_PEA,
		templatePath:   "constraints/PodEnforceAntiaffinity/template/pod-enforce-antiaffinity.yaml",
		constraintPath: "constraints/PodEnforceAntiaffinity/constraint/constraint.yaml",
	},
	{
		name:           Constraint_RT,
		templatePath:   "constraints/RestrictedTaints/template/restricted-taints.yaml",
		constraintPath: "constraints/RestrictedTaints/constraint/constraint.yaml",
	},
	{
		name:           Constraint_USS,
		templatePath:   "constraints/UniqueServiceSelectors/template/unique-service-selectors.yaml",
		constraintPath: "constraints/UniqueServiceSelectors/constraint/constraint.yaml",
	},
}

func getConstraintClient() (*constraintclient.Client, error) {
	driver, err := rego.New()
	if err != nil {
		return nil, fmt.Errorf("could not create rego driver: %w", err.Error())
	}

	c, err := constraintclient.NewClient(constraintclient.Targets(&target.K8sValidationTarget{}), constraintclient.Driver(driver))
	if err != nil {
		return nil, fmt.Errorf("could not create constraint client: %w", err)
	}

	return c, nil
}

// primes the scheme to be able to interpret beta templates
func init() {
	_ = clientgoscheme.AddToScheme(s)
	_ = api.AddToScheme(s)
}

func (fc FileCrawler) ReadDeployment(path string) (*unstructured.Unstructured, error) {
	deployment, err := reader.ReadObject(f, path)
	if err != nil {
		return nil, fmt.Errorf("could not read deployment: %w", err.Error())
	}

	return deployment, nil
}

func (fc FileCrawler) ReadConstraintTemplates() ([]*templates.ConstraintTemplate, error) {
	var constraintTemplates []*templates.ConstraintTemplate

	for _, sg := range safeguards {
		ct, err := reader.ReadTemplate(s, f, sg.templatePath)
		if err != nil {
			return nil, fmt.Errorf("could not read template: %w", err.Error())
		}
		constraintTemplates = append(constraintTemplates, ct)
	}

	return constraintTemplates, nil
}

func (fc FileCrawler) ReadConstraintTemplate(name string) (*templates.ConstraintTemplate, error) {
	var constraintTemplate *templates.ConstraintTemplate

	for _, sg := range safeguards {
		if sg.name == name {
			ct, err := reader.ReadTemplate(s, f, sg.templatePath)
			if err != nil {
				return nil, fmt.Errorf("could not read template: %w", err.Error())
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

	for _, sg := range safeguards {
		u, err := reader.ReadConstraint(f, sg.constraintPath)
		if err != nil {
			return nil, fmt.Errorf("could not add constraint: %w", err.Error())
		}

		constraints = append(constraints, u)
	}

	return constraints, nil
}

func (fc FileCrawler) ReadConstraint(name string) (*unstructured.Unstructured, error) {
	var constraint *unstructured.Unstructured

	for _, sg := range safeguards {
		if sg.name == name {
			c, err := reader.ReadConstraint(f, sg.constraintPath)
			if err != nil {
				return nil, fmt.Errorf("could not add constraint: %w", err.Error())
			}

			constraint = c
		}
	}
	if constraint == nil {
		return nil, fmt.Errorf("no constraint exists with name: %s", name)
	}

	return constraint, nil
}

func loadConstraintTemplates(ctx context.Context, c *constraintclient.Client, constraintTemplates []*templates.ConstraintTemplate) error {
	// AddTemplate adds the template source code to OPA and registers the CRD with the client for
	// schema validation on calls to AddConstraint. On error, the responses return value
	// will still be populated so that partial results can be analyzed.
	for _, ct := range constraintTemplates {
		_, err := c.AddTemplate(ctx, ct)
		if err != nil {
			return fmt.Errorf("could not add template: %w", err.Error())
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
			return fmt.Errorf("could not add constraint: %w", err.Error())
		}
	}

	return nil
}

func validateDeployment(ctx context.Context, c *constraintclient.Client, deployment *unstructured.Unstructured) error {
	// Review makes sure the provided object satisfies all stored constraints.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.
	res, err := c.Review(ctx, deployment)
	if err != nil {
		return fmt.Errorf("could not review deployment: %w", err.Error())
	}

	for _, v := range res.ByTarget {
		for _, result := range v.Results {
			if result.Msg != "" {
				return fmt.Errorf("deployment error: %s", result.Msg)
			}
		}
	}

	return nil
}

// ValidateDeployment is what will be called by `draft validate` to validate the user's deployment manifest
// against each safeguards constraint
func ValidateDeployment(ctx context.Context, deploymentPath string) error {
	var fc FileCrawler

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
	deployment, err := fc.ReadDeployment(deploymentPath)
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
	return validateDeployment(ctx, c, deployment)
}
