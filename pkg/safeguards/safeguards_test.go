package safeguards

import (
	"context"
	"testing"

	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ctx = context.Background()
var fc FileCrawler

func init() {
	fc = FileCrawler{
		Safeguards: safeguards,
	}
}

func TestValidateSafeguardsConstraint_CAI(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := fc.ReadConstraintTemplate(testDeployment_CAI.Name)
	assert.Nil(t, err)
	constraint, err := fc.ReadConstraint(testDeployment_CAI.Name)
	assert.Nil(t, err)
	errDeployment, err := fc.ReadDeployment(testDeployment_CAI.ErrorPath)
	assert.Nil(t, err)
	successDeployment, err := fc.ReadDeployment(testDeployment_CAI.SuccessPath)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	// error case - should throw error
	err = validateDeployment(ctx, c, errDeployment)
	assert.NotNil(t, err)
	// success case - should not throw error
	err = validateDeployment(ctx, c, successDeployment)
	assert.Nil(t, err)
}

// TODO: investigate whether or not to include more than one success/error case
//
//	for deployments being tested
func TestValidateSafeguardsConstraint_CEP(t *testing.T) {
	// instantiate constraint client
	c, err := getConstraintClient()
	assert.Nil(t, err)

	// retrieving template, constraint, and deployments
	constraintTemplate, err := fc.ReadConstraintTemplate(testDeployment_CEP.Name)
	assert.Nil(t, err)
	constraint, err := fc.ReadConstraint(testDeployment_CEP.Name)
	assert.Nil(t, err)
	errDeployment, err := fc.ReadDeployment(testDeployment_CEP.ErrorPath)
	assert.Nil(t, err)
	successDeployment, err := fc.ReadDeployment(testDeployment_CEP.SuccessPath)
	assert.Nil(t, err)

	// load template, constraint into constraint client
	err = loadConstraintTemplates(ctx, c, []*templates.ConstraintTemplate{constraintTemplate})
	assert.Nil(t, err)
	err = loadConstraints(ctx, c, []*unstructured.Unstructured{constraint})
	assert.Nil(t, err)

	// validating deployment manifests
	// error case - should throw error
	err = validateDeployment(ctx, c, errDeployment)
	assert.NotNil(t, err)
	// success case - should not throw error
	err = validateDeployment(ctx, c, successDeployment)
	assert.Nil(t, err)
}
