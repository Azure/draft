package preprocessing

import (
	"bytes"
	"testing"

	consts "github.com/Azure/draft/pkg/safeguards/types"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Test rendering a valid Helm chart with no subcharts and three templates
func TestRenderHelmChart_Valid(t *testing.T) {
	var opt chartutil.ReleaseOptions

	manifestFiles, err := RenderHelmChart(false, consts.ChartPath, opt)
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
	manifestFiles, err = RenderHelmChart(true, consts.DirectPath_ToValidChart, opt)
	assert.Nil(t, err)

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}
}

// Test rendering a valid Helm chart with no subcharts and three templates, using command line flags
func TestRenderHelmChartWithFlags_Valid(t *testing.T) {
	// user defined release name and namespace from cli flags
	opt := chartutil.ReleaseOptions{
		Name:      "test-flags-name",
		Namespace: "test-flags-namespace",
	}

	manifestFiles, err := RenderHelmChart(false, consts.ChartPath, opt)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := make(map[string][]byte)
	expectedFiles["deployment.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expecteddeployment_flags.yaml")
	expectedFiles["service.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expectedservice_flags.yaml")
	expectedFiles["ingress.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expectedingress_flags.yaml")

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}

	// Test by giving file directly
	manifestFiles, err = RenderHelmChart(true, consts.DirectPath_ToValidChart, opt)
	assert.Nil(t, err)

	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
	}
}

// Should successfully render a Helm chart with sub charts and be able to render subchart separately within a helm chart
func TestSubCharts(t *testing.T) {
	var opt chartutil.ReleaseOptions

	manifestFiles, err := RenderHelmChart(false, consts.Subcharts, opt)
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
	_, err = RenderHelmChart(false, consts.SubchartDir, opt)
	assert.Nil(t, err)

	// Given a Chart.yaml in the main directory, main chart and subcharts should be evaluated
	_, err = RenderHelmChart(true, consts.SubchartDir, opt)
	assert.Nil(t, err)

	// Given path to a sub-Chart.yaml with a dependency on another subchart, should render both subcharts, but not the main chart
	manifestFiles, err = RenderHelmChart(true, consts.DirectPath_ToSubchartYaml, opt)
	assert.Nil(t, err)

	expectedFiles = make(map[string][]byte)
	expectedFiles["deployment1.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart1.yaml")
	expectedFiles["deployment2.yaml"] = getManifestAsBytes(t, "../tests/testmanifests/expected-subchart2.yaml")

	assert.Equal(t, len(manifestFiles), 2)
	for i, writtenManifestFile := range manifestFiles {
		writtenFileName := manifestFiles[i].Name
		expectedYaml := bytes.TrimSpace(expectedFiles[writtenFileName])
		assert.Equal(t, bytes.TrimSpace(writtenManifestFile.ManifestContent), expectedYaml)
		assert.NoFileExists(t, "maindeployment.yaml", "Unexpected file was created: maindeployment.yaml")
	}
}

/**
* Testing user errors
 */

// Should fail if the Chart.yaml is invalid
func TestInvalidChartAndValues(t *testing.T) {
	var opt chartutil.ReleaseOptions

	_, err := RenderHelmChart(false, consts.InvalidChartPath, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load main chart: validation: chart.metadata.name is required")

	_, err = RenderHelmChart(false, consts.InvalidValuesChart, opt)
	assert.NotNil(t, err)
}

// Testing with malformed Deployment.yaml
func TestInvalidDeployments(t *testing.T) {
	var opt chartutil.ReleaseOptions

	_, err := RenderHelmChart(false, consts.InvalidDeploymentSyntax, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")

	_, err = RenderHelmChart(false, consts.InvalidDeploymentValues, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "map has no entry for key")
}

// Test different helm folder structures
func TestDifferentFolderStructures(t *testing.T) {
	var opt chartutil.ReleaseOptions
	manifestFiles, err := RenderHelmChart(false, consts.FolderwithHelpersTmpl, opt) // includes _helpers.tpl
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

	manifestFiles, err = RenderHelmChart(false, consts.MultipleTemplateDirs, opt) // all manifests defined in one file
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
	_, err := RenderKustomizeManifest(consts.KustomizationPath)
	assert.Nil(t, err)
}
