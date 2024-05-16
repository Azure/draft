package safeguards

import (
	"fmt"
)

const (
	Constraint_CAI   = "container-allowed-images"
	Constraint_CEP   = "container-enforce-probes"
	Constraint_CRL   = "container-resource-limits"
	Constraint_CRIP  = "container-restricted-image-pulls"
	Constraint_DBPDB = "disallowed-bad-pod-disruption-budgets"
	Constraint_PEA   = "pod-enforce-antiaffinity"
	Constraint_RT    = "restricted-taints"
	Constraint_USS   = "unique-service-selectors"
	Constraint_all   = "all"
)

var selectedVersion = "v1.0.0"

// TODO: consider getting this from a text file we can bump
var supportedVersions = []string{selectedVersion}

const (
	templateFileName   = "template.yaml"
	constraintFileName = "constraint.yaml"
)

var Safeguard_CRIP = Safeguard{
	name:           Constraint_CRIP,
	templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CRIP, templateFileName),
	constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CRIP, constraintFileName),
}

var safeguards = []Safeguard{
	{
		name:           Constraint_CAI,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CAI, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CAI, constraintFileName),
	},
	{
		name:           Constraint_CEP,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CEP, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CEP, constraintFileName),
	},
	{
		name:           Constraint_CRL,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CRL, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_CRL, constraintFileName),
	},
	{
		name:           Constraint_DBPDB,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_DBPDB, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_DBPDB, constraintFileName),
	},
	{
		name:           Constraint_PEA,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_PEA, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_PEA, constraintFileName),
	},
	{
		name:           Constraint_RT,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_RT, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_RT, constraintFileName),
	},
	{
		name:           Constraint_USS,
		templatePath:   fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_USS, templateFileName),
		constraintPath: fmt.Sprintf("lib/%s/%s/%s", selectedVersion, Constraint_USS, constraintFileName),
	},
}

var safeguardsTesting = append(safeguards, Safeguard_CRIP, Safeguard_CAI)
