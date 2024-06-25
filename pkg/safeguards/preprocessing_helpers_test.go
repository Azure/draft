package safeguards

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	tempDir                  = "testdata" // Rendered files are stored here before they are read for comparison
	chartPath                = "tests/testmanifests/validchart"
	invalidChartPath         = "tests/testmanifests/invalidchart"
	invalidValuesChart       = "tests/testmanifests/invalidvalues"
	invalidDeploymentsChart  = "tests/testmanifests/invaliddeployment"
	invalidDeploymentSyntax  = "tests/testmanifests/invaliddeployment-syntax"
	invalidDeploymentValues  = "tests/testmanifests/invaliddeployment-values"
	differentFolderStructure = "tests/testmanifests/different-structure"
	multipleTemplateDirs     = "tests/testmanifests/multiple-templates"
	multipleValuesFile       = "tests/testmanifests/multiple-values-files"
	subcharts                = "tests/testmanifests/multiple-charts"
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

func TestRenderHelmChart_Valid(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(chartPath, tempDir)
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
}

func TestSubCharts(t *testing.T) {
	setup(t)
	//t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(subcharts, tempDir)
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
}

/**
* Testing user errors
 */

// Should fail if the Chart.yaml is invalid
func TestInvalidChart(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(invalidChartPath, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load main chart: validation: chart.metadata.name is required")
}

// Should fail if values.yaml doesn't contain all values necessary for templating
func TestInvalidValues(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })
	err := RenderHelmChart(invalidValuesChart, tempDir)
	assert.NotNil(t, err)
}

func TestInvalidDeployments(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(invalidDeploymentSyntax, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")

	err = RenderHelmChart(invalidDeploymentValues, tempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "map has no entry for key")
}

func TestDifferentFolderStructure(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(differentFolderStructure, tempDir)
	assert.Nil(t, err)
}

// Tests both multiple sub directories and multiple resources defined in one file
func TestResourcesInOneFile(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(multipleTemplateDirs, tempDir)
	assert.Nil(t, err)
}

func TestMutlipleValuesFiles(t *testing.T) {
	setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := RenderHelmChart(multipleValuesFile, tempDir)
	assert.Nil(t, err)
}

func cleanupDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Printf("Failed to clean directory: %s", err)

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
