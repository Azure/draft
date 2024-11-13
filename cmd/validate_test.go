package cmd

import (
	"context"
	"testing"

	"github.com/Azure/draft/pkg/safeguards"
	"github.com/Azure/draft/pkg/safeguards/preprocessing"
	"github.com/Azure/draft/pkg/safeguards/types"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/chartutil"
)

const (
	manifestPathDirectorySuccess = "../pkg/safeguards/tests/all/success"
	manifestPathDirectoryError   = "../pkg/safeguards/tests/all/error"
	manifestPathFileSuccess      = "../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml"
	manifestPathFileError        = "../pkg/safeguards/tests/all/error/all-error-manifest-1.yaml"
	kustomizationPath            = "../pkg/safeguards/tests/kustomize/overlays/production"
	chartPath                    = "../pkg/safeguards/tests/testmanifests/validchart"
	kustomizationFilePath        = "../pkg/safeguards/tests/kustomize/overlays/production/kustomization.yaml"
)

// TestRunValidate tests the run command for `draft validate` for proper returns
func TestRunValidate(t *testing.T) {
	ctx := context.TODO()
	manifestFilesEmpty := []types.ManifestFile{}

	var manifestFiles []types.ManifestFile
	var opt chartutil.ReleaseOptions

	// Scenario 1: empty manifest path should error
	_, err := safeguards.GetManifestResults(ctx, manifestFilesEmpty)
	assert.NotNil(t, err)

	// Scenario 2a: manifest path leads to a directory of manifestFiles - expect success
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathDirectorySuccess, opt)
	assert.Nil(t, err)
	v, err := safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations := countTestViolations(v)
	assert.Equal(t, numViolations, 0)

	// Scenario 2b: manifest path leads to a directory of manifestFiles - expect failure
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathDirectoryError, opt)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)

	// Scenario 3a: manifest path leads to one manifest file - expect success
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathFileSuccess, opt)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Equal(t, numViolations, 0)

	// Scenario 3b: manifest path leads to one manifest file - expect failure
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathFileError, opt)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)

	//Scenario 4: Test Kustomize
	manifestFiles, err = safeguards.GetManifestFiles(kustomizationPath, opt)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)

	// Scenario 5: Test Helm
	opt.Name = "test-release-name"
	opt.Namespace = "test-release-namespace"

	manifestFiles, err = safeguards.GetManifestFiles(chartPath, opt)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)
}

// TestRunValidate_Kustomize tests the run command for `draft validate` for proper returns when given a kustomize project
func TestRunValidate_Kustomize(t *testing.T) {
	ctx := context.TODO()
	var manifestFiles []types.ManifestFile
	var err error

	// Scenario 1a: kustomizationPath leads to a directory containing kustomization.yaml - expect success
	manifestFiles, err = preprocessing.RenderKustomizeManifest(kustomizationPath)
	assert.Nil(t, err)
	v, err := safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations := countTestViolations(v)
	assert.Equal(t, numViolations, 1)

	// Scenario 1b: kustomizationFilePath path leads to a specific kustomization.yaml - expect success
	manifestFiles, err = preprocessing.RenderKustomizeManifest(kustomizationFilePath)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Equal(t, numViolations, 1)
}
