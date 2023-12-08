package safeguards

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testDeployment struct {
	ConstraintName string
	SuccessPath    string
	ErrorPath      string
}

var testDeployment_CAI = testDeployment{
	ConstraintName: Constraint_CAI,
	SuccessPath:    "constraints/ContainerAllowedImages/testcases/CAI_Success_Manifest.yaml",
	ErrorPath:      "constraints/ContainerAllowedImages/testcases/CAI_Error_Manifest.yaml",
}

var testDeployment_CEP = testDeployment{
	ConstraintName: Constraint_CEP,
	SuccessPath:    "constraints/ContainerEnforceProbes/testcases/CEP_Success_Manifest.yaml",
	ErrorPath:      "constraints/ContainerEnforceProbes/testcases/CEP_Error_Manifest.yaml",
}

var testDeployment_CRL = testDeployment{
	ConstraintName: Constraint_CRL,
	SuccessPath:    "constraints/ContainerResourceLimits/testcases/CRL_Success_Manifest.yaml",
	ErrorPath:      "constraints/ContainerResourceLimits/testcases/CRL_Error_Manifest.yaml",
}

var testDeployment_NUP = testDeployment{
	ConstraintName: Constraint_NUP,
	SuccessPath:    "constraints/NoUnauthenticatedPulls/testcases/NUP_Success_Manifest.yaml",
	ErrorPath:      "constraints/NoUnauthenticatedPulls/testcases/NUP_Error_Manifest.yaml",
}

var testDeployment_PDB = testDeployment{
	ConstraintName: Constraint_PDB,
	SuccessPath:    "constraints/PodDisruptionBudgets/testcases/PDB_Success_Manifest.yaml",
	ErrorPath:      "constraints/PodDisruptionBudgets/testcases/PDB_Error_Manifest.yaml",
}

var testDeployment_PEA = testDeployment{
	ConstraintName: Constraint_PEA,
	SuccessPath:    "constraints/PodEnforceAntiaffinity/testcases/PEA_Success_Manifest.yaml",
	ErrorPath:      "constraints/PodEnforceAntiaffinity/testcases/PEA_Error_Manifest.yaml",
}

var testDeployment_RT = testDeployment{
	ConstraintName: Constraint_RT,
	SuccessPath:    "constraints/RestrictedTaints/testcases/RT_Success_Manifest.yaml",
	ErrorPath:      "constraints/RestrictedTaints/testcases/RT_Error_Manifest.yaml",
}

var testDeployment_USS = testDeployment{
	ConstraintName: Constraint_USS,
	SuccessPath:    "constraints/UniqueServiceSelectors/testcases/USS_Success_Manifest.yaml",
	ErrorPath:      "constraints/UniqueServiceSelectors/testcases/USS_Error_Manifest.yaml",
}

// thbarnes: working on error case(s); investigate if default can just be used for all success
func TestValidateSafeguardsConstraint_CAI(t *testing.T) {
	ctx := context.Background()
	var fcf FilesystemConstraintFetcher
	constraintFile, err := fcf.FetchOne(testDeployment_CAI.ConstraintName)
	assert.Nil(t, err)

	deployment, err := fetchDeploymentFile(testDeployment_CAI.ErrorPath)
	assert.Nil(t, err)

	err = evaluateQuery(ctx, constraintFile, deployment)
	assert.NotNil(t, err)

	deployment, err = fetchDeploymentFile(testDeployment_CAI.SuccessPath)
	assert.Nil(t, err)

	err = evaluateQuery(ctx, constraintFile, deployment)
	assert.Nil(t, err)
}
