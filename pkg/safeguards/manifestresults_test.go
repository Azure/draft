package safeguards

import (
	"context"
	"testing"

	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ctx = context.Background()
var testFc FileCrawler

func init() {
	selectedVersion = getLatestSafeguardsVersion()
	updateSafeguardPaths(&safeguardsTesting)

	testFc = FileCrawler{
		Safeguards:   safeguardsTesting,
		constraintFS: embedFS,
	}
}

// TestValidateDeployment_ContainerAllowedImages tests our Container Allowed Images manifest
func TestValidateDeployment_ContainerAllowedImages(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_CAI.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_CAI.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	// tbarnes94: disabling failure case until we finalize logic on imageRegex for allowed images
	//validateOneTestManifestFail(ctx, t, c, testFc, testManifest_CAI.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_CAI.SuccessPaths)
}

// TestValidateDeployment_ContainerEnforceProbes tests our Container Enforce Probes manifest
func TestValidateDeployment_ContainerEnforceProbes(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_CEP.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_CEP.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_CEP.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_CEP.SuccessPaths)
}

// TestValidateDeployment_ContainerResourceLimits tests our Container Resource Limits manifest
func TestValidateDeployment_ContainerResourceLimits(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_CL.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_CL.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_CL.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_CL.SuccessPaths)
}

// TestValidateDeployment_ContainerRestrictedImagePulls tests our Container Restricted Image Pulls manifest
func TestValidateDeployment_ContainerRestrictedImagePulls(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_CRIP.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_CRIP.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_CRIP.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_CRIP.SuccessPaths)
}

// TestValidateDeployment_DisallowedBadPodDisruptionBudget tests our Disallowed Bad Pod Disruption Budget manifest
func TestValidateDeployment_DisallowedBadPodDisruptionBudget(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_DBPDB.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_DBPDB.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_DBPDB.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_DBPDB.SuccessPaths)
}

// TestValidateDeployment_PodEnforceAntiaffinity tests our Pod Enforce Antiaffinity manifest
func TestValidateDeployment_PodEnforceAntiaffinity(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_PEA.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_PEA.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_PEA.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_PEA.SuccessPaths)
}

// TestValidateDeployment_RestrictedTaints tests our Restricted Taints manifest
func TestValidateDeployment_RestrictedTaints(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_RT.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_RT.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_RT.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_RT.SuccessPaths)
}

// TestValidateDeployment_UniqueServiceSelectors tests our Unique Service Selectors manifest
func TestValidateDeployment_UniqueServiceSelectors(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_USS.Name)
	assert.Nil(t, err)
	constraint, err := testFc.ReadConstraint(testManifest_USS.Name)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	validateOneTestManifestFail(ctx, t, c, testFc, testManifest_USS.ErrorPaths)
	validateOneTestManifestSuccess(ctx, t, c, testFc, testManifest_USS.SuccessPaths)
}

// TestValidateDeployment_All tests all of our manifests in a few given manifest files
func TestValidateDeployment_All(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	var testTemplates []*templates.ConstraintTemplate
	var testConstraints []*unstructured.Unstructured
	for _, sg := range safeguards {
		// retrieving template, constraint, and deployments
		constraintTemplate, err := testFc.ReadConstraintTemplate(sg.name)
		assert.Nil(t, err)
		testTemplates = append(testTemplates, constraintTemplate)

		constraint, err := testFc.ReadConstraint(sg.name)
		assert.Nil(t, err)
		testConstraints = append(testConstraints, constraint)

	}

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, testTemplates)
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, testConstraints)
	assert.Nil(t, err)

	// validating deployment manifests
	validateAllTestManifestsFail(ctx, t, c, testFc, testManifest_all.ErrorPaths)
	validateAllTestManifestsSuccess(ctx, t, c, testFc, testManifest_all.SuccessPaths)
}
