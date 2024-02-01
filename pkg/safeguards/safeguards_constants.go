package safeguards

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
