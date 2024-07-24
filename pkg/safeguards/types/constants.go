package types

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

var SelectedVersion = "v1.0.0"

var SupportedVersions = []string{SelectedVersion}

const (
	TemplateFileName   = "template.yaml"
	ConstraintFileName = "constraint.yaml"
)

var Safeguard_CRIP = Safeguard{
	Name:           Constraint_CRIP,
	TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CRIP, TemplateFileName),
	ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CRIP, ConstraintFileName),
}

var Safeguards = []Safeguard{
	{
		Name:           Constraint_CAI,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CAI, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CAI, ConstraintFileName),
	},
	{
		Name:           Constraint_CEP,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CEP, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CEP, ConstraintFileName),
	},
	{
		Name:           Constraint_CRL,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CRL, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_CRL, ConstraintFileName),
	},
	{
		Name:           Constraint_DBPDB,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_DBPDB, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_DBPDB, ConstraintFileName),
	},
	{
		Name:           Constraint_PEA,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_PEA, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_PEA, ConstraintFileName),
	},
	{
		Name:           Constraint_RT,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_RT, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_RT, ConstraintFileName),
	},
	{
		Name:           Constraint_USS,
		TemplatePath:   fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_USS, TemplateFileName),
		ConstraintPath: fmt.Sprintf("lib/%s/%s/%s", SelectedVersion, Constraint_USS, ConstraintFileName),
	},
}

var SafeguardsTesting = append(Safeguards, Safeguard_CRIP)
