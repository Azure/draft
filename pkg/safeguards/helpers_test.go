package safeguards

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/draft/pkg/safeguards/types"
	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	chartPath               = "tests/testmanifests/validchart"
	kustomizationPath       = "tests/kustomize/overlays/production"
	kustomizationFilePath   = "tests/kustomize/overlays/production/kustomization.yaml"
	directPath_ToValidChart = "tests/testmanifests/validchart/Chart.yaml"
)

func TestGetManifestFiles(t *testing.T) {
	var opt chartutil.ReleaseOptions
	// Test Helm
	_, err := GetManifestFiles(chartPath, opt)
	assert.Nil(t, err)

	// Test Kustomize
	_, err = GetManifestFiles(kustomizationPath, opt)
	assert.Nil(t, err)

	// Test Normal Directory with manifest files
	absPath, err := filepath.Abs("tests/all/success")
	assert.Nil(t, err)
	_, err = GetManifestFiles(absPath, opt)
	assert.Nil(t, err)

	// test single manifest file
	manifestPathFileSuccess, err := filepath.Abs("tests/all/success/all-success-manifest-1.yaml")
	assert.Nil(t, err)
	_, err = GetManifestFiles(manifestPathFileSuccess, opt)
	assert.Nil(t, err)
}

func validateOneTestManifestFail(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc types.FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		byteContent, err := os.ReadFile(path)
		assert.Nil(t, err)

		errManifests, err := testFc.ReadManifests(byteContent)
		assert.Nil(t, err)

		err = loadManifestObjects(ctx, c, errManifests)
		assert.Nil(t, err)

		// error case - should throw error
		violations, err := getObjectViolations(ctx, c, errManifests)
		assert.Nil(t, err)
		assert.Greater(t, len(violations), 0)
	}
}

func validateOneTestManifestSuccess(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc types.FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		byteContent, err := os.ReadFile(path)
		assert.Nil(t, err)

		successManifests, err := testFc.ReadManifests(byteContent)
		assert.Nil(t, err)

		err = loadManifestObjects(ctx, c, successManifests)
		assert.Nil(t, err)

		// success case - should not throw error
		violations, err := getObjectViolations(ctx, c, successManifests)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(violations))
	}
}

func validateAllTestManifestsFail(ctx context.Context, t *testing.T, testManifestPaths []string) {
	var opt chartutil.ReleaseOptions
	for _, path := range testManifestPaths {
		manifestFiles, err := GetManifestFiles(path, opt)
		assert.Nil(t, err)

		// error case - should throw error
		results, err := GetManifestResults(ctx, manifestFiles)
		for _, mr := range results {
			assert.Nil(t, err)
			assert.Greater(t, mr.ViolationsCount, 0)
		}
	}
}

func validateAllTestManifestsSuccess(ctx context.Context, t *testing.T, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		manifestFiles, err := GetManifestFilesFromDir(path)
		assert.Nil(t, err)

		// success case - should not throw error
		results, err := GetManifestResults(ctx, manifestFiles)
		for _, mr := range results {
			assert.Nil(t, err)
			assert.Equal(t, mr.ViolationsCount, 0)
		}
	}
}

// TestIsKustomize checks whether the given path contains a kustomize project
func TestIsKustomize(t *testing.T) {
	kustomizationPath := "tests/kustomize/overlays/production"

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
