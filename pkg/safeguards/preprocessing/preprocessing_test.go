package preprocessing

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test rendering a valid Helm chart with no subcharts and three templates
func TestRenderHelmChart_Valid(t *testing.T) {
	manifestFiles, err := RenderHelmChart(false, chartPath)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := make(map[string][]byte)
	expectedFiles["deployment.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expecteddeployment.yaml")
	expectedFiles["service.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expectedservice.yaml")
	expectedFiles["ingress.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expectedingress.yaml")

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}

	// Test by giving file directly
	manifestFiles, err = RenderHelmChart(true, directPath_ToValidChart)
	assert.Nil(t, err)

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}
}

// Should successfully render a Helm chart with sub charts and be able to render subchart separately within a helm chart
func TestSubCharts(t *testing.T) {
	manifestFiles, err := RenderHelmChart(false, subcharts)
	assert.Nil(t, err)

	expectedFiles := make(map[string][]byte)
	expectedFiles["maindeployment.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-mainchart.yaml")
	expectedFiles["deployment1.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart1.yaml")
	expectedFiles["deployment2.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart2.yaml")

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}

	// Given a sub-chart dir, that specific sub chart only should be evaluated and rendered
	_, err = RenderHelmChart(false, subchartDir)
	assert.Nil(t, err)

	// Given a Chart.yaml in the main directory, main chart and subcharts should be evaluated
	_, err = RenderHelmChart(true, directPath_ToMainChartYaml)
	assert.Nil(t, err)

	// Given path to a sub-Chart.yaml with a dependency on another subchart, should render both subcharts, but not the main chart
	manifestFiles, err = RenderHelmChart(true, directPath_ToSubchartYaml)
	assert.Nil(t, err)

	expectedFiles = make(map[string][]byte)
	expectedFiles["deployment1.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart1.yaml")
	expectedFiles["deployment2.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart2.yaml")

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}

	//expect mainchart.yaml to not exist
	outputFilePath := filepath.Join(tempDir, "maindeployment.yaml")
	assert.NoFileExists(t, outputFilePath, "Unexpected file was created: %s", outputFilePath)
}

/**
* Testing user errors
 */

// Should fail if the Chart.yaml is invalid
func TestInvalidChartAndValues(t *testing.T) {
	_, err := RenderHelmChart(false, invalidChartPath)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load main chart: validation: chart.metadata.name is required")

	_, err = RenderHelmChart(false, invalidValuesChart)
	assert.NotNil(t, err)
}

func TestInvalidDeployments(t *testing.T) {
	_, err := RenderHelmChart(false, invalidDeploymentSyntax)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")

	_, err = RenderHelmChart(false, invalidDeploymentValues)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "map has no entry for key")
}

// Test different helm folder structures
func TestDifferentFolderStructures(t *testing.T) {
	manifestFiles, err := RenderHelmChart(false, folderwithHelpersTmpl) // includes _helpers.tpl
	assert.Nil(t, err)

	expectedFiles := make(map[string][]byte)
	expectedFiles["deployment.yaml"] = normalizeNewlines(getManifestAsBytes(t, "../tests/testmanifests/expected-helpers-deployment.yaml"))
	expectedFiles["service.yaml"] = normalizeNewlines(getManifestAsBytes(t, "../tests/testmanifests/expected-helpers-service.yaml"))
	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		resFile := bytes.TrimSpace(normalizeNewlines(writtenManifestFile.ManifestContent))
		assert.Equal(t, resFile, expectedYaml)
	}

	manifestFiles, err = RenderHelmChart(false, multipleTemplateDirs) // all manifests defined in one file
	assert.Nil(t, err)

	expectedFiles = make(map[string][]byte)
	expectedFiles["resources.yaml"] = normalizeNewlines(getManifestAsBytes(t, "../tests/testmanifests/expected-resources.yaml"))
	expectedFiles["service-1.yaml"] = normalizeNewlines(getManifestAsBytes(t, "../tests/testmanifests/expectedservice.yaml"))
	expectedFiles["service-2.yaml"] = normalizeNewlines(getManifestAsBytes(t, "../tests/testmanifests/expectedservice2.yaml"))
	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		resFile := bytes.TrimSpace(normalizeNewlines(writtenManifestFile.ManifestContent))
		assert.Equal(t, string(resFile), string(expectedYaml))
	}
}

// Test rendering a valid kustomization.yaml
func TestRenderKustomizeManifest_Valid(t *testing.T) {
	_, err := RenderKustomizeManifest(kustomizationPath)
	assert.Nil(t, err)
}

// TODO: later update these tests to validate the file content returned in the ManifestFile struct
func TestGetManifestFiles(t *testing.T) {
	// Test Helm
	_, err := GetManifestFiles(chartPath)
	assert.Nil(t, err)
	_, err = GetManifestFiles(kustomizationPath)
	assert.Nil(t, err)

	// Test Kustomize
	_, err = GetManifestFiles(kustomizationPath)
	assert.Nil(t, err)

	// Test Normal Directory with manifest files
	absPath, err := filepath.Abs("../tests/all/success")
	assert.Nil(t, err)
	_, err = GetManifestFiles(absPath)
	assert.Nil(t, err)

	// test single manifest file
	manifestPathFileSuccess, err := filepath.Abs("../tests/all/success/all-success-manifest-1.yaml")
	assert.Nil(t, err)
	_, err = GetManifestFiles(manifestPathFileSuccess)
	assert.Nil(t, err)
}

// TestIsKustomize checks whether the given path contains a kustomize project
func TestIsKustomize(t *testing.T) {
	// path contains a kustomization.yaml file
	iskustomize := isKustomize(true, kustomizationPath)
	assert.True(t, iskustomize)
	// path is a kustomization.yaml file
	iskustomize = isKustomize(false, kustomizationFilePath)
	assert.True(t, iskustomize)
	// not a kustomize project
	iskustomize = isKustomize(true, chartPath)
	assert.False(t, iskustomize)
}

func TestIsHelm(t *testing.T) {
	// path is a directory
	ishelm := isHelm(true, chartPath)
	assert.True(t, ishelm)

	// path is a Chart.yaml file
	ishelm = isHelm(false, directPath_ToValidChart)
	assert.True(t, ishelm)

	// Is a directory but does not contain Chart.yaml
	ishelm = isHelm(true, kustomizationPath)
	assert.False(t, ishelm)

	// Is a directory of manifest files, not a helm chart
	ishelm = isHelm(false, "../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")
	assert.False(t, ishelm)

	// Is a directory of manifest files, not a helm chart
	ishelm = isHelm(false, "../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")
	assert.False(t, ishelm)

	// invalid path
	ishelm = isHelm(false, "invalid/path")
	assert.False(t, ishelm)
}
