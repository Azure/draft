package preprocessing

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	chartPath               = "../tests/testmanifests/validchart"
	invalidChartPath        = "../tests/testmanifests/invalidchart"
	invalidValuesChart      = "../tests/testmanifests/invalidvalues"
	invalidDeploymentsChart = "../tests/testmanifests/invaliddeployment"
	invalidDeploymentSyntax = "../tests/testmanifests/invaliddeployment-syntax"
	invalidDeploymentValues = "../tests/testmanifests/invaliddeployment-values"
	folderwithHelpersTmpl   = "../tests/testmanifests/different-structure"
	multipleTemplateDirs    = "../tests/testmanifests/multiple-templates"
	multipleValuesFile      = "../tests/testmanifests/multiple-values-files"

	subcharts                  = "../tests/testmanifests/multiple-charts"
	subchartDir                = "../tests/testmanifests/multiple-charts/charts/subchart2"
	directPath_ToSubchartYaml  = "../tests/testmanifests/multiple-charts/charts/subchart1/Chart.yaml"
	directPath_ToMainChartYaml = "../tests/testmanifests/multiple-charts/Chart.yaml"

	directPath_ToValidChart   = "../tests/testmanifests/validchart/Chart.yaml"
	directPath_ToInvalidChart = "../tests/testmanifests/invalidchart/Chart.yaml"

	kustomizationPath     = "../tests/kustomize/overlays/production"
	kustomizationFilePath = "../tests/kustomize/overlays/production/kustomization.yaml"
)

func getManifestAsBytes(t *testing.T, filePath string) []byte {
	yamlFileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %s", err)
	}

	return yamlFileContent
}

// replace newlines with strings for easy .yaml byte comparison
func normalizeNewlines(data []byte) []byte {
	str := string(data)

	// Replace newlines and carriage returns with a single newline
	str = strings.ReplaceAll(str, "\r\n", "\n")
	str = strings.ReplaceAll(str, "\r", "\n")

	// Replace YAML block scalars' indicators with empty space
	// Handles cases like "data: config.yaml: |"
	re := regexp.MustCompile(`(\s*\|\s*)`)
	str = re.ReplaceAllString(str, " ")

	// Replace multiple spaces with a single space
	str = strings.Join(strings.Fields(str), " ")

	// Normalize empty mappings and fields
	str = regexp.MustCompile(`(\{\s*\})`).ReplaceAllString(str, "{}")
	str = regexp.MustCompile(`(\s*:\s*)`).ReplaceAllString(str, ": ")

	return []byte(str)
}
