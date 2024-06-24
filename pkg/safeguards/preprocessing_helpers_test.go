package safeguards

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	testManifestsDir = "testmanifests"
	tempDir          = "testdata"
	chartPath        = "testdata/my-web-app"             // Path to the test chart directory
	valuesPath       = "testdata/my-web-app/values.yaml" // Path to the test values.yaml file
	outputDir        = "testdata/output/manifests"       // Path to the output directory
)

type testVars struct {
	validChartYaml       string
	validValuesYaml      string
	validDeploymentYaml  string
	validServiceYaml     string
	validIngressYaml     string
	invalidChartYaml     string
	invalidValuesYaml    string
	invalidTemplateYamls map[string]string
}

/*
* Tests to cover:
* Invalid values.yaml
* Invalid chart.yaml -- what is the bare minimum needed for a chart.yaml? What fields, if included, would break this function?
* short template files, long template files.
* One template file, multiple template files. One valid, others invalid
 */

func setup(t *testing.T) testVars {
	// Ensure the output directory is empty before running the test
	if err := os.RemoveAll(outputDir); err != nil {
		t.Fatalf("Failed to clean output directory: %s", err)
	}

	// Create the templates directory
	err := os.MkdirAll(filepath.Join(chartPath, "templates"), 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %s", err)
	}

	//validTemplateYamls: map[string]string{"service.yaml": serviceYAML}
	chartYAML := getManifestAsString("validchart.yaml")
	valuesYAML := getManifestAsString("validvalues.yaml")
	validDeploymentYaml := getManifestAsString("validdeployment.yaml.tmpl")
	validServiceYaml := getManifestAsString("validservice.yaml.tmpl")
	validIngressYaml := getManifestAsString("validingress.yaml.tmpl")
	invalidChartYaml := getManifestAsString("invalidchart.yaml")
	invalidValuesYaml := getManifestAsString("invalidvalues.yaml")

	// Create invalid templates
	invalidTemplateYamls := make(map[string]string)
	invalidTemplateYamls["deployment.yaml"] = getManifestAsString("invaliddeployment.yaml.tmpl")
	invalidTemplateYamls["deploymentSyntax.yaml"] = getManifestAsString("invaliddeploymentsyntax.yaml.tmpl")
	invalidTemplateYamls["deploymentValues.yaml"] = getManifestAsString("invaliddeploymentvalues.yaml.tmpl")

	return testVars{validChartYaml: chartYAML, validValuesYaml: valuesYAML, validDeploymentYaml: validDeploymentYaml, validServiceYaml: validServiceYaml, validIngressYaml: validIngressYaml, invalidChartYaml: invalidChartYaml, invalidValuesYaml: invalidValuesYaml, invalidTemplateYamls: invalidTemplateYamls}
}

func TestRenderHelmChart_Valid(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = os.WriteFile(valuesPath, []byte(v.validValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.validDeploymentYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/service.yaml"), []byte(v.validServiceYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/service.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/ingress.yaml"), []byte(v.validIngressYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/ingress.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, outputDir)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := []string{"deployment.yaml", "service.yaml", "ingress.yaml"}
	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(outputDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file does not exist: %s", outputFilePath)
	}

	//assert that each file output matches expected yaml after values are filled in
	assert.Equal(t, parseYAML(t, getManifestAsString("expecteddeployment.yaml")), parseYAML(t, readFile(t, filepath.Join(outputDir, "deployment.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString("expectedservice.yaml")), parseYAML(t, readFile(t, filepath.Join(outputDir, "service.yaml"))))
	assert.Equal(t, parseYAML(t, getManifestAsString("expectedingress.yaml")), parseYAML(t, readFile(t, filepath.Join(outputDir, "ingress.yaml"))))

}

/**
* Testing user errors
 */

// Should fail if the chart and values.yaml are invalid
func TestInvalidChartAndValues(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	// Invalid Chart.yaml
	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.invalidChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = RenderHelmChart(chartPath, valuesPath, outputDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load chart: validation: chart.metadata.name is required")

	// Write valid Chart.yaml
	err = os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	// Write invalid values.yaml
	err = os.WriteFile(filepath.Join(chartPath, "values.yaml"), []byte(v.invalidValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	// Write the service template
	err = os.WriteFile(filepath.Join(chartPath, "templates/service.yaml"), []byte(v.validServiceYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/service.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, outputDir)

	// Assert that an error occurs
	assert.NotNil(t, err)
	fmt.Println("err: ", err)
	assert.Contains(t, err.Error(), "failed to render chart:")
}

func TestInvalidTemplate(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(tempDir) })

	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = os.WriteFile(valuesPath, []byte(v.validValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.invalidTemplateYamls["deploymentValues.yaml"]), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, outputDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to render chart: template: my-web-app/templates/deployment.yaml")
	assert.Contains(t, err.Error(), "map has no entry for key \"nonExistentField\"")

	cleanupDir(outputDir)
	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.invalidTemplateYamls["deploymentSyntax.yaml"]), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, outputDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")
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

func getManifestAsString(filename string) string {
	yamlFilePath := filepath.Join("tests/testmanifests", filename)

	yamlFileContent, err := os.ReadFile(yamlFilePath)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %s", err)
	}

	// Convert the content to a string
	yamlContentString := string(yamlFileContent)
	return yamlContentString
}
