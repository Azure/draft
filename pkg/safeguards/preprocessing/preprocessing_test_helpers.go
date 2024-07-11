package preprocessing

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	tempDir                 = "testdata" // Rendered files are stored here before they are read for comparison
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

	kustomizationPath = "../tests/kustomize/overlays/production"
)

func makeTempDir(t *testing.T) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temporary output directory: %s", err)
	}
}

func cleanupDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to clean directory: %s", err)
	}
}

func parseYAML(t *testing.T, content string) map[string]interface{} {
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %s", err)
	}
	return result
}

func getManifestAsString(t *testing.T, filePath string) string {
	yamlFileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %s", err)
	}

	yamlContentString := string(yamlFileContent)
	return yamlContentString
}
