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

	KustomizationPath       = "../tests/kustomize/overlays/production"
	DirectPath_ToValidChart = "../tests/testmanifests/validchart/Chart.yaml"
	ChartPath               = "../tests/testmanifests/validchart"
	InvalidChartPath        = "../tests/testmanifests/invalidchart"
	InvalidValuesChart      = "../tests/testmanifests/invalidvalues"
	InvalidDeploymentSyntax = "../tests/testmanifests/invaliddeployment-syntax"
	InvalidDeploymentValues = "../tests/testmanifests/invaliddeployment-values"
	FolderwithHelpersTmpl   = "../tests/testmanifests/different-structure"
	MultipleTemplateDirs    = "../tests/testmanifests/multiple-templates"
	MultipleValuesFile      = "../tests/testmanifests/multiple-values-files"

	Subcharts                  = "../tests/testmanifests/multiple-charts"
	SubchartDir                = "../tests/testmanifests/multiple-charts/charts/subchart2"
	DirectPath_ToSubchartYaml  = "../tests/testmanifests/multiple-charts/charts/subchart1/Chart.yaml"
	DirectPath_ToMainChartYaml = "../tests/testmanifests/multiple-charts/Chart.yaml"
	DirectPath_ToInvalidChart  = "../tests/testmanifests/invalidchart/Chart.yaml"

	TemplateFileName   = "template.yaml"
	ConstraintFileName = "constraint.yaml"
)

var SelectedVersion = "v1.0.0"

var SupportedVersions = []string{SelectedVersion}

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
