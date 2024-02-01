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

type FileCrawler struct {
	Safeguards []Safeguard
}

// primes the scheme to be able to interpret beta templates
func init() {
	_ = clientgoscheme.AddToScheme(s)
	_ = api.AddToScheme(s)
}

// ValidateDeployment is what will be called by `draft validate` to validate the user's deployment manifest
// against each safeguards constraint
func ValidateDeployment(ctx context.Context, deploymentPath string) error {
	fc := FileCrawler{
		Safeguards: safeguards,
	}

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
