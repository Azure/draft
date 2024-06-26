package safeguards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	tempDir                 = "testdata" // Rendered files are stored here before they are read for comparison
	chartPath               = "tests/testmanifests/validchart"
	invalidChartPath        = "tests/testmanifests/invalidchart"
	invalidValuesChart      = "tests/testmanifests/invalidvalues"
	invalidDeploymentsChart = "tests/testmanifests/invaliddeployment"
	invalidDeploymentSyntax = "tests/testmanifests/invaliddeployment-syntax"
	invalidDeploymentValues = "tests/testmanifests/invaliddeployment-values"
	folderwithHelpersTmpl   = "tests/testmanifests/different-structure"
	multipleTemplateDirs    = "tests/testmanifests/multiple-templates"
	multipleValuesFile      = "tests/testmanifests/multiple-values-files"

	subcharts                  = "tests/testmanifests/multiple-charts"
	subchartDir                = "tests/testmanifests/multiple-charts/charts/subchart2"
	directPath_ToSubchartYaml  = "tests/testmanifests/multiple-charts/charts/subchart1/Chart.yaml"
	directPath_ToMainChartYaml = "tests/testmanifests/multiple-charts/Chart.yaml"

	directPath_ToValidChart   = "tests/testmanifests/validchart/Chart.yaml"
	directPath_ToInvalidChart = "tests/testmanifests/invalidchart/Chart.yaml"
)

func setup(t *testing.T) {
	// Ensure the output directory is empty before running the test
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to clean output directory: %s", err)
	}

	// Create the templates directory
	err := os.MkdirAll(filepath.Join(chartPath, "templates"), 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %s", err)
	}
}

// Test rendering a valid Helm chart with no subcharts and three templates
func TestRenderHelmChart_Valid(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	err := RenderHelmChart(false, chartPath, tempDir)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := []string{"deployment.yaml", "service.yaml", "ingress.yaml"}
	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(tempDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file was not created: %s", outputFilePath)
	}

	//assert that each file output matches expected yaml after values are filled in
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expecteddeployment.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expectedservice.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "service.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expectedingress.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "ingress.yaml"))))

	cleanupDir(t, tempDir)

	//Test by giving file directly
	err = RenderHelmChart(true, directPath_ToValidChart, tempDir)
	assert.Nil(t, err)

	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(tempDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file was not created: %s", outputFilePath)
	}

	//assert that each file output matches expected yaml after values are filled in
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expecteddeployment.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expectedservice.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "service.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expectedingress.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "ingress.yaml"))))
}

// Should successfully render a Helm chart with sub charts and be able to render subchart separately within a helm chart
func TestSubCharts(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	err := RenderHelmChart(false, subcharts, tempDir)
	assert.Nil(t, err)

	//assert that 3 files were created in temp dir: 1 from main chart, 2 from subcharts
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, len(files), 3)
	expectedFiles := []string{"maindeployment.yaml", "deployment1.yaml", "deployment2.yaml"}
	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(tempDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file was not created: %s", outputFilePath)
	}
	//assert that the files are equal
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expected-mainchart.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "maindeployment.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expected-subchart1.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment1.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expected-subchart2.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment2.yaml"))))

	cleanupDir(t, tempDir)

	// Given a sub-chart dir, that specific sub chart only should be evaluated and rendered
	err = RenderHelmChart(false, subchartDir, tempDir)
	assert.Nil(t, err)

	cleanupDir(t, tempDir)

	// Given a Chart.yaml in the main directory, main chart and subcharts should be evaluated
	err = RenderHelmChart(true, directPath_ToMainChartYaml, tempDir)
	assert.Nil(t, err)

	cleanupDir(t, tempDir)

	//Given path to a sub- Chart.yaml with a dependency on another subchart, should render both subcharts, but not the main chart
	err = RenderHelmChart(true, directPath_ToSubchartYaml, tempDir)
	assert.Nil(t, err)
	expectedFiles = []string{"deployment1.yaml", "deployment2.yaml"}
	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(tempDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file was not created: %s", outputFilePath)
	}
	//expect mainchart.yaml to not exist
	outputFilePath := filepath.Join(tempDir, "maindeployment.yaml")
	assert.NoFileExists(t, outputFilePath, "Unexpected file was created: %s", outputFilePath)
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expected-subchart1.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment1.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString(t, "expected-subchart2.yaml")), parseYAML(t, readFile(t, filepath.Join(tempDir, "deployment2.yaml"))))
}

/**
* Testing user errors
 */

// Should fail if the Chart.yaml is invalid
func TestInvalidChartAndValues(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	err := RenderHelmChart(false, invalidChartPath, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load main chart: validation: chart.metadata.name is required")

	err = RenderHelmChart(true, directPath_ToValidChart, tempDir)
	assert.Nil(t, err)

	// Should fail if values.yaml doesn't contain all values necessary for templating
	cleanupDir(t, tempDir)
	err = RenderHelmChart(false, invalidValuesChart, tempDir)
	assert.NotNil(t, err)
}

func TestInvalidDeployments(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	err := RenderHelmChart(false, invalidDeploymentSyntax, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")

	err = RenderHelmChart(false, invalidDeploymentValues, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "map has no entry for key")
}

/** Test different helm folder structures */
func TestDifferentFolderStructures(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	err := RenderHelmChart(false, folderwithHelpersTmpl, tempDir) // includes _helpers.tpl
	assert.Nil(t, err)

	cleanupDir(t, tempDir)
	err = RenderHelmChart(false, multipleTemplateDirs, tempDir) // all manifests defined in one file
	assert.Nil(t, err)

	cleanupDir(t, tempDir)
	err = RenderHelmChart(false, multipleValuesFile, tempDir) // contains three values files
	assert.Nil(t, err)
}

func cleanupDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to clean directory: %s", err)
	}
}

func readFile(t *testing.T, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %s", err)
	}
	return string(data)
}

func parseYAML(t *testing.T, content string) map[string]interface{} {
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %s", err)
	}
	return result
}

func getManifestAsString(t *testing.T, filename string) string {
	yamlFilePath := filepath.Join("tests/testmanifests", filename)

	yamlFileContent, err := os.ReadFile(yamlFilePath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %s", err)
	}

	yamlContentString := string(yamlFileContent)
	return yamlContentString
}
