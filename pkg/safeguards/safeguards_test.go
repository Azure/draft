package safeguards

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
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

type testDeployment struct {
	ConstraintName string
	SuccessPath    string
	ErrorPath      string
}

var testDeployment_CAI = testDeployment{
	ConstraintName: Constraint_CAI,
	SuccessPath:    "pkg/safeguards/constraints/ContainerAllowedImages/testcases/CAI_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/ContainerAllowedImages/testcases/CAI_Error_Manifest.yaml",
}

var testDeployment_CEP = testDeployment{
	ConstraintName: Constraint_CEP,
	SuccessPath:    "pkg/safeguards/constraints/ContainerEnforceProbes/testcases/CEP_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/ContainerEnforceProbes/testcases/CEP_Error_Manifest.yaml",
}

var testDeployment_CRL = testDeployment{
	ConstraintName: Constraint_CRL,
	SuccessPath:    "pkg/safeguards/constraints/ContainerResourceLimits/testcases/CRL_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/ContainerResourceLimits/testcases/CRL_Error_Manifest.yaml",
}

var testDeployment_NUP = testDeployment{
	ConstraintName: Constraint_NUP,
	SuccessPath:    "pkg/safeguards/constraints/NoUnauthenticatedPulls/testcases/NUP_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/NoUnauthenticatedPulls/testcases/NUP_Error_Manifest.yaml",
}

var testDeployment_PDB = testDeployment{
	ConstraintName: Constraint_PDB,
	SuccessPath:    "pkg/safeguards/constraints/PodDisruptionBudgets/testcases/PDB_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/PodDisruptionBudgets/testcases/PDB_Error_Manifest.yaml",
}

var testDeployment_PEA = testDeployment{
	ConstraintName: Constraint_PEA,
	SuccessPath:    "pkg/safeguards/constraints/PodEnforceAntiaffinity/testcases/PEA_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/PodEnforceAntiaffinity/testcases/PEA_Error_Manifest.yaml",
}

var testDeployment_RT = testDeployment{
	ConstraintName: Constraint_RT,
	SuccessPath:    "pkg/safeguards/constraints/RestrictedTaints/testcases/RT_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/RestrictedTaints/testcases/RT_Error_Manifest.yaml",
}

var testDeployment_USS = testDeployment{
	ConstraintName: Constraint_USS,
	SuccessPath:    "pkg/safeguards/constraints/UniqueServiceSelectors/testcases/USS_Success_Manifest.yaml",
	ErrorPath:      "pkg/safeguards/constraints/UniqueServiceSelectors/testcases/USS_Error_Manifest.yaml",
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

	err = ValidateDeployment(testDeployment_CAI.SuccessPath, testDeployment_CAI.ConstraintName)
	assert.Nil(t, err)
}
