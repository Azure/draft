package preprocessing

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/chartutil"
)

// Test rendering a valid Helm chart with no subcharts and three templates
func TestRenderHelmChart_Valid(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	var opt chartutil.ReleaseOptions

	manifestFiles, err := RenderHelmChart(false, chartPath, tempDir, opt)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := make(map[string]string)
	expectedFiles["deployment.yaml"] = getManifestAsString(t, "../tests/testmanifests/expecteddeployment.yaml")
	expectedFiles["service.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedservice.yaml")
	expectedFiles["ingress.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedingress.yaml")

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}

	cleanupDir(t, tempDir)
	makeTempDir(t)

	// Test by giving file directly
	manifestFiles, err = RenderHelmChart(true, directPath_ToValidChart, tempDir, opt)
	assert.Nil(t, err)

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}
}

// Test rendering a valid Helm chart with no subcharts and three templates, using command line flags
func TestRenderHelmChartWithFlags_Valid(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	// user defined release name and namespace from cli flags
	opt := chartutil.ReleaseOptions{
		Name:      "test-flags-name",
		Namespace: "test-flags-namespace",
	}

	manifestFiles, err := RenderHelmChart(false, chartPath, tempDir, opt)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := make(map[string]string)
	expectedFiles["deployment.yaml"] = getManifestAsString(t, "../tests/testmanifests/expecteddeployment_flags.yaml")
	expectedFiles["service.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedservice_flags.yaml")
	expectedFiles["ingress.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedingress_flags.yaml")

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}

	cleanupDir(t, tempDir)
	makeTempDir(t)

	// Test by giving file directly
	manifestFiles, err = RenderHelmChart(true, directPath_ToValidChart, tempDir, opt)
	assert.Nil(t, err)

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}
}

// Should successfully render a Helm chart with sub charts and be able to render subchart separately within a helm chart
func TestSubCharts(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	var opt chartutil.ReleaseOptions

	manifestFiles, err := RenderHelmChart(false, subcharts, tempDir, opt)
	assert.Nil(t, err)

	// Assert that 3 files were created in temp dir: 1 from main chart, 2 from subcharts
	files, _ := os.ReadDir(tempDir)
	assert.Equal(t, len(files), 3)

	expectedFiles := make(map[string]string)
	expectedFiles["maindeployment.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-mainchart.yaml")
	expectedFiles["deployment1.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-subchart1.yaml")
	expectedFiles["deployment2.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-subchart2.yaml")

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}

	cleanupDir(t, tempDir)
	makeTempDir(t)

	// Given a sub-chart dir, that specific sub chart only should be evaluated and rendered
	_, err = RenderHelmChart(false, subchartDir, tempDir, opt)
	assert.Nil(t, err)

	cleanupDir(t, tempDir)
	makeTempDir(t)

	// Given a Chart.yaml in the main directory, main chart and subcharts should be evaluated
	_, err = RenderHelmChart(true, directPath_ToMainChartYaml, tempDir, opt)
	assert.Nil(t, err)

	cleanupDir(t, tempDir)
	makeTempDir(t)

	// Given path to a sub-Chart.yaml with a dependency on another subchart, should render both subcharts, but not the main chart
	manifestFiles, err = RenderHelmChart(true, directPath_ToSubchartYaml, tempDir, opt)
	assert.Nil(t, err)

	expectedFiles = make(map[string]string)
	expectedFiles["deployment1.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-subchart1.yaml")
	expectedFiles["deployment2.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-subchart2.yaml")

	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
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
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	var opt chartutil.ReleaseOptions

	_, err := RenderHelmChart(false, invalidChartPath, tempDir, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load main chart: validation: chart.metadata.name is required")

	_, err = RenderHelmChart(true, directPath_ToValidChart, tempDir, opt)
	assert.Nil(t, err)

	// Should fail if values.yaml doesn't contain all values necessary for templating
	cleanupDir(t, tempDir)
	makeTempDir(t)

	_, err = RenderHelmChart(false, invalidValuesChart, tempDir, opt)
	assert.NotNil(t, err)
}

func TestInvalidDeployments(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	var opt chartutil.ReleaseOptions

	_, err := RenderHelmChart(false, invalidDeploymentSyntax, tempDir, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")

	_, err = RenderHelmChart(false, invalidDeploymentValues, tempDir, opt)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "map has no entry for key")
}

// Test different helm folder structures
func TestDifferentFolderStructures(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })
	var opt chartutil.ReleaseOptions

	manifestFiles, err := RenderHelmChart(false, folderwithHelpersTmpl, tempDir, opt) // includes _helpers.tpl
	assert.Nil(t, err)

	expectedFiles := make(map[string]string)
	expectedFiles["deployment.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-helpers-deployment.yaml")
	expectedFiles["service.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-helpers-service.yaml")
	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}
	cleanupDir(t, tempDir)
	makeTempDir(t)

	manifestFiles, err = RenderHelmChart(false, multipleTemplateDirs, tempDir, opt) // all manifests defined in one file
	assert.Nil(t, err)

	expectedFiles = make(map[string]string)
	expectedFiles["resources.yaml"] = getManifestAsString(t, "../tests/testmanifests/expected-resources.yaml")
	expectedFiles["service-1.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedservice.yaml")
	expectedFiles["service-2.yaml"] = getManifestAsString(t, "../tests/testmanifests/expectedservice2.yaml")
	for _, writtenFile := range manifestFiles {
		expectedYaml := expectedFiles[writtenFile.Name]
		writtenYaml := parseYAML(t, getManifestAsString(t, writtenFile.Path))
		assert.Equal(t, writtenYaml, parseYAML(t, expectedYaml))
	}
}

// Test rendering a valid kustomization.yaml
func TestRenderKustomizeManifest_Valid(t *testing.T) {
	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	_, err := RenderKustomizeManifest(kustomizationPath, tempDir)
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

// TestIsYAML tests the IsYAML function for proper returns
func TestIsYAML(t *testing.T) {
	dirNotYaml, _ := filepath.Abs("../tests/not-yaml")
	dirYaml, _ := filepath.Abs("../tests/all/success")
	fileNotYaml, _ := filepath.Abs("../tests/not-yaml/readme.md")
	fileYaml, _ := filepath.Abs("/tests/all/success/all-success-manifest-1.yaml")
	var opt chartutil.ReleaseOptions

	assert.False(t, IsYAML(fileNotYaml))
	assert.True(t, IsYAML(fileYaml))

	manifestFiles, err := GetManifestFiles(dirNotYaml, opt)
	assert.Nil(t, manifestFiles)
	assert.NotNil(t, err)

	manifestFiles, err = GetManifestFiles(dirYaml, opt)
	assert.NotNil(t, manifestFiles)
	assert.Nil(t, err)
}

// TestIsDirectory tests the isDirectory function for proper returns
func TestIsDirectory(t *testing.T) {
	testWd, _ := os.Getwd()
	pathTrue := testWd
	pathFalse := path.Join(testWd, "preprocessing.go")
	pathError := ""

	isDir, err := IsDirectory(pathTrue)
	assert.True(t, isDir)
	assert.Nil(t, err)

	isDir, err = IsDirectory(pathFalse)
	assert.False(t, isDir)
	assert.Nil(t, err)

	isDir, err = IsDirectory(pathError)
	assert.False(t, isDir)
	assert.NotNil(t, err)
}
