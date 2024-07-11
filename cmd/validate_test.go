package cmd

import (
	"context"
	"github.com/Azure/draft/pkg/safeguards/preprocessing"
	"os"
	"path"
	"path/filepath"

	"testing"

	"github.com/Azure/draft/pkg/safeguards"
	"github.com/stretchr/testify/assert"
)

// TestIsDirectory tests the isDirectory function for proper returns
func TestIsDirectory(t *testing.T) {
	testWd, _ := os.Getwd()
	pathTrue := testWd
	pathFalse := path.Join(testWd, "validate.go")
	pathError := ""

	isDir, err := safeguards.IsDirectory(pathTrue)
	assert.True(t, isDir)
	assert.Nil(t, err)

	isDir, err = safeguards.IsDirectory(pathFalse)
	assert.False(t, isDir)
	assert.Nil(t, err)

	isDir, err = safeguards.IsDirectory(pathError)
	assert.False(t, isDir)
	assert.NotNil(t, err)
}

// TestIsYAML tests the isYAML function for proper returns
func TestIsYAML(t *testing.T) {
	dirNotYaml, _ := filepath.Abs("../pkg/safeguards/tests/not-yaml")
	dirYaml, _ := filepath.Abs("../pkg/safeguards/tests/all/success")
	fileNotYaml, _ := filepath.Abs("../pkg/safeguards/tests/not-yaml/readme.md")
	fileYaml, _ := filepath.Abs("../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")

	assert.False(t, safeguards.IsYAML(fileNotYaml))
	assert.True(t, safeguards.IsYAML(fileYaml))

	manifestFiles, err := safeguards.GetManifestFiles(dirNotYaml)
	assert.Nil(t, manifestFiles)
	assert.NotNil(t, err)

	manifestFiles, err = safeguards.GetManifestFiles(dirYaml)
	assert.NotNil(t, manifestFiles)
	assert.Nil(t, err)
}

// TestRunValidate tests the run command for `draft validate` for proper returns
func TestRunValidate(t *testing.T) {
	ctx := context.TODO()
	manifestFilesEmpty := []safeguards.ManifestFile{}
	manifestPathDirectorySuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success")
	manifestPathDirectoryError, _ := filepath.Abs("../pkg/safeguards/tests/all/error")
	manifestPathFileSuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")
	manifestPathFileError, _ := filepath.Abs("../pkg/safeguards/tests/all/error/all-error-manifest-1.yaml")
	var manifestFiles []safeguards.ManifestFile

	// Scenario 1: empty manifest path should error
	_, err := safeguards.GetManifestResults(ctx, manifestFilesEmpty)
	assert.NotNil(t, err)

	// Scenario 2a: manifest path leads to a directory of manifestFiles - expect success
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathDirectorySuccess)
	assert.Nil(t, err)
	v, err := safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations := countTestViolations(v)
	assert.Equal(t, numViolations, 0)

	// Scenario 2b: manifest path leads to a directory of manifestFiles - expect failure
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathDirectoryError)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)

	// Scenario 3a: manifest path leads to one manifest file - expect success
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathFileSuccess)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Equal(t, numViolations, 0)

	// Scenario 3b: manifest path leads to one manifest file - expect failure
	manifestFiles, err = safeguards.GetManifestFiles(manifestPathFileError)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Greater(t, numViolations, 0)
}

// TestRunValidate_Kustomize tests the run command for `draft validate` for proper returns when given a kustomize project
func TestRunValidate_Kustomize(t *testing.T) {
	ctx := context.TODO()
	kustomizationPath, _ := filepath.Abs("../pkg/safeguards/tests/kustomize/overlays/production")
	kustomizationFilePath, _ := filepath.Abs("../pkg/safeguards/tests/kustomize/overlays/production/kustomization.yaml")

	makeTempDir(t)
	t.Cleanup(func() { cleanupDir(t, tempDir) })

	var manifestFiles []safeguards.ManifestFile
	var err error

	// Scenario 1a: kustomizationPath leads to a directory containing kustomization.yaml - expect success
	manifestFiles, err = preprocessing.RenderKustomizeManifest(kustomizationPath, tempDir)
	assert.Nil(t, err)
	v, err := safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations := countTestViolations(v)
	assert.Equal(t, numViolations, 1)

	// Scenario 1b: kustomizationFilePath path leads to a specific kustomization.yaml - expect success
	manifestFiles, err = preprocessing.RenderKustomizeManifest(kustomizationFilePath, tempDir)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestResults(ctx, manifestFiles)
	assert.Nil(t, err)
	numViolations = countTestViolations(v)
	assert.Equal(t, numViolations, 1)
}
