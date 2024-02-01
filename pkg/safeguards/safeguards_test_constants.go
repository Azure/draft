package safeguards

type testDeployment struct {
	Name        string
	SuccessPath string
	ErrorPath   string
}

var testDeployment_CAI = testDeployment{
	Name:        Constraint_CAI,
	SuccessPath: "constraints/ContainerAllowedImages/deployments/CAI_Success_Manifest.yaml",
	ErrorPath:   "constraints/ContainerAllowedImages/deployments/CAI_Error_Manifest.yaml",
}

var testDeployment_CEP = testDeployment{
	Name:        Constraint_CEP,
	SuccessPath: "constraints/ContainerEnforceProbes/deployments/CEP_Success_Manifest.yaml",
	ErrorPath:   "constraints/ContainerEnforceProbes/deployments/CEP_Error_Manifest.yaml",
}

var testDeployment_CRL = testDeployment{
	Name:        Constraint_CRL,
	SuccessPath: "constraints/ContainerResourceLimits/deployments/CRL_Success_Manifest.yaml",
	ErrorPath:   "constraints/ContainerResourceLimits/deployments/CRL_Error_Manifest.yaml",
}

var testDeployment_NUP = testDeployment{
	Name:        Constraint_NUP,
	SuccessPath: "constraints/NoUnauthenticatedPulls/deployments/NUP_Success_Manifest.yaml",
	ErrorPath:   "constraints/NoUnauthenticatedPulls/deployments/NUP_Error_Manifest.yaml",
}

var testDeployment_PDB = testDeployment{
	Name:        Constraint_PDB,
	SuccessPath: "constraints/PodDisruptionBudgets/deployments/PDB_Success_Manifest.yaml",
	ErrorPath:   "constraints/PodDisruptionBudgets/deployments/PDB_Error_Manifest.yaml",
}

var testDeployment_PEA = testDeployment{
	Name:        Constraint_PEA,
	SuccessPath: "constraints/PodEnforceAntiaffinity/deployments/PEA_Success_Manifest.yaml",
	ErrorPath:   "constraints/PodEnforceAntiaffinity/deployments/PEA_Error_Manifest.yaml",
}

var testDeployment_RT = testDeployment{
	Name:        Constraint_RT,
	SuccessPath: "constraints/RestrictedTaints/deployments/RT_Success_Manifest.yaml",
	ErrorPath:   "constraints/RestrictedTaints/deployments/RT_Error_Manifest.yaml",
}

var testDeployment_USS = testDeployment{
	Name:        Constraint_USS,
	SuccessPath: "constraints/UniqueServiceSelectors/deployments/USS_Success_Manifest.yaml",
	ErrorPath:   "constraints/UniqueServiceSelectors/deployments/USS_Error_Manifest.yaml",
}
