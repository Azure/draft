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
	updateSafeguardPaths()

	testFc = FileCrawler{
		Safeguards: safeguards,
	}
}

// TODO: rich description here
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
	validateTestManifests_Error(ctx, t, c, testFc, testManifest_CAI.ErrorPaths)
	validateTestManifests_Success(ctx, t, c, testFc, testManifest_CAI.SuccessPaths)
}

// TODO: rich description here
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
	validateTestManifests_Error(ctx, t, c, testFc, testManifest_CEP.ErrorPaths)
	validateTestManifests_Success(ctx, t, c, testFc, testManifest_CEP.SuccessPaths)
}

// // TODO: rich description here
func TestValidateDeployment_ContainerLimits(t *testing.T) {
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
	validateTestManifests_Error(ctx, t, c, testFc, testManifest_CL.ErrorPaths)
	validateTestManifests_Success(ctx, t, c, testFc, testManifest_CL.SuccessPaths)
}

//// TODO: rich description here
//func TestValidateDeployment_ContainerRestrictedImagePulls(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	// retrieving template, constraint, and deployments
//	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_CRIP.Name)
//	assert.Nil(t, err)
//	constraint, err := testFc.ReadConstraint(testManifest_CRIP.Name)
//	assert.Nil(t, err)
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_CRIP.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_CRIP.SuccessPaths)
//}
//
//// TODO: rich description here
//func TestValidateDeployment_DisallowedBadPodDisruptionBudget(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	// retrieving template, constraint, and deployments
//	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_DBPDB.Name)
//	assert.Nil(t, err)
//	constraint, err := testFc.ReadConstraint(testManifest_DBPDB.Name)
//	assert.Nil(t, err)
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_DBPDB.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_DBPDB.SuccessPaths)
//}
//
//// TODO: rich description here
//func TestValidateDeployment_PodEnforceAntiaffinity(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	// retrieving template, constraint, and deployments
//	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_PEA.Name)
//	assert.Nil(t, err)
//	constraint, err := testFc.ReadConstraint(testManifest_PEA.Name)
//	assert.Nil(t, err)
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_PEA.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_PEA.SuccessPaths)
//}
//
//// TODO: rich description here
//func TestValidateDeployment_RestrictedTaints(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	// retrieving template, constraint, and deployments
//	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_RT.Name)
//	assert.Nil(t, err)
//	constraint, err := testFc.ReadConstraint(testManifest_RT.Name)
//	assert.Nil(t, err)
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_RT.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_RT.SuccessPaths)
//}
//
//// TODO: rich description here
//func TestValidateDeployment_UniqueServiceSelectors(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	// retrieving template, constraint, and deployments
//	constraintTemplate, err := testFc.ReadConstraintTemplate(testManifest_USS.Name)
//	assert.Nil(t, err)
//	constraint, err := testFc.ReadConstraint(testManifest_USS.Name)
//	assert.Nil(t, err)
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_USS.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_USS.SuccessPaths)
//}
//
//// TODO: rich description here
//func TestValidateDeployment_All(t *testing.T) {
//	// instantiate constraint client
//	c, err := getConstraintClient()
//	assert.Nil(t, err)
//
//	var testTemplates []*templates.ConstraintTemplate
//	var testConstraints []*unstructured.Unstructured
//	for _, sg := range safeguards {
//		// retrieving template, constraint, and deployments
//		constraintTemplate, err := testFc.ReadConstraintTemplate(sg.name)
//		assert.Nil(t, err)
//		testTemplates = append(testTemplates, constraintTemplate)
//
//		constraint, err := testFc.ReadConstraint(sg.name)
//		assert.Nil(t, err)
//		testConstraints = append(testConstraints, constraint)
//
//	}
//
//	// load template, constraint into constraint client
//	err = loadConstraintTemplates(ctx, c, testTemplates)
//	assert.Nil(t, err)
//	err = loadConstraints(ctx, c, testConstraints)
//	assert.Nil(t, err)
//
//	// validating deployment manifests
//	validateTestManifests_Error(ctx, t, c, testFc, testManifest_all.ErrorPaths)
//	validateTestManifests_Success(ctx, t, c, testFc, testManifest_all.SuccessPaths)
//}
