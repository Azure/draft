package safeguards

import (
	"context"
	"testing"

	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Azure/draft/pkg/safeguards/types"
)

var ctx = context.Background()
var testFc types.FileCrawler

func init() {
	types.SelectedVersion = getLatestSafeguardsVersion()
	updateSafeguardPaths(&types.SafeguardsTesting)

	testFc = types.FileCrawler{
		Safeguards:   types.SafeguardsTesting,
		ConstraintFS: embedFS,
	}
}

// TestValidateDeployment_ContainerAllowedImages tests our Container Allowed Images manifest
func TestValidateDeployment_ContainerAllowedImages(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_CAI.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_CAI.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	// tbarnes94: disabling failure case until we finalize logic on imageRegex for allowed images
	//validateOneTestManifestFail(ctx, t, c, testFc, h.TestManifest_CAI.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_CAI.SuccessPaths)
}

// TestValidateDeployment_ContainerEnforceProbes tests our Container Enforce Probes manifest
func TestValidateDeployment_ContainerEnforceProbes(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_CEP.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_CEP.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_CEP.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_CEP.SuccessPaths)
}

// TestValidateDeployment_ContainerResourceLimits tests our Container Resource Limits manifest
func TestValidateDeployment_ContainerResourceLimits(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_CL.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_CL.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_CL.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_CL.SuccessPaths)
}

// TestValidateDeployment_ContainerRestrictedImagePulls tests our Container Restricted Image Pulls manifest
func TestValidateDeployment_ContainerRestrictedImagePulls(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_CRIP.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_CRIP.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_CRIP.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_CRIP.SuccessPaths)
}

// TestValidateDeployment_DisallowedBadPodDisruptionBudget tests our Disallowed Bad Pod Disruption Budget manifest
func TestValidateDeployment_DisallowedBadPodDisruptionBudget(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_DBPDB.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_DBPDB.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_DBPDB.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_DBPDB.SuccessPaths)
}

// TestValidateDeployment_PodEnforceAntiaffinity tests our Pod Enforce Antiaffinity manifest
func TestValidateDeployment_PodEnforceAntiaffinity(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_PEA.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_PEA.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_PEA.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_PEA.SuccessPaths)
}

// TestValidateDeployment_RestrictedTaints tests our Restricted Taints manifest
func TestValidateDeployment_RestrictedTaints(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_RT.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_RT.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_RT.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_RT.SuccessPaths)
}

// TestValidateDeployment_UniqueServiceSelectors tests our Unique Service Selectors manifest
func TestValidateDeployment_UniqueServiceSelectors(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(types.TestManifest_USS.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(types.TestManifest_USS.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, types.TestManifest_USS.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, types.TestManifest_USS.SuccessPaths)
}

// TestValidateDeployment_All tests all of our manifests in a few given manifest files
func TestValidateDeployment_All(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	var testTemplates []*templates.ConstraintTemplate
	var testConstraints []*unstructured.Unstructured
	for _, sg := range types.Safeguards {
		// retrieving template, constraint, and deployments
		constraintTemplate, err := testFc.ReadConstraintTemplate(sg.Name)
		assert.Nil(t, err)
		testTemplates = append(testTemplates, constraintTemplate)

		constraint, err := testFc.ReadConstraint(sg.Name)
		assert.Nil(t, err)
		testConstraints = append(testConstraints, constraint)

	}

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, testTemplates)
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, testConstraints)
	assert.Nil(t, err)

	// validating deployment manifests
	validateAllTestManifestsFail(ctx, t, types.TestManifest_all.ErrorPaths)
	validateAllTestManifestsSuccess(ctx, t, types.TestManifest_all.SuccessPaths)
}
