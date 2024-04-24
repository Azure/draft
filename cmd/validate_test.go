package cmd

import (
	"context"
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

	isDir, err := isDirectory(pathTrue)
	assert.True(t, isDir)
	assert.Nil(t, err)

	isDir, err = isDirectory(pathFalse)
	assert.False(t, isDir)
	assert.Nil(t, err)

	isDir, err = isDirectory(pathError)
	assert.False(t, isDir)
	assert.NotNil(t, err)
}

// TestIsYAML tests the isYAML function for proper returns
func TestIsYAML(t *testing.T) {
	dirNotYaml, _ := filepath.Abs("../pkg/safeguards/tests/not-yaml")
	dirYaml, _ := filepath.Abs("../pkg/safeguards/tests/all/success")
	fileNotYaml, _ := filepath.Abs("../pkg/safeguards/tests/not-yaml/readme.md")
	fileYaml, _ := filepath.Abs("../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")

	assert.False(t, isYAML(fileNotYaml))
	assert.True(t, isYAML(fileYaml))

	manifestFiles, err := getManifestFiles(dirNotYaml)
	assert.Nil(t, manifestFiles)
	assert.NotNil(t, err)

	manifestFiles, err = getManifestFiles(dirYaml)
	assert.NotNil(t, manifestFiles)
	assert.Nil(t, err)
}

// TestRunValidate tests the run command for `draft validate` for proper returns
func TestRunValidate(t *testing.T) {
	ctx := context.TODO()
	manifestPathEmpty := ""
	manifestPathDirectorySuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success")
	manifestPathDirectoryError, _ := filepath.Abs("../pkg/safeguards/tests/all/error")
	manifestPathFileSuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success/all-success-manifest-1.yaml")
	manifestPathFileError, _ := filepath.Abs("../pkg/safeguards/tests/all/error/all-error-manifest-1.yaml")
	var manifestFiles []string

	// Scenario 1: empty manifest path should error
	manifestFiles = append(manifestFiles, manifestPathEmpty)
	_, err := safeguards.GetManifestViolations(ctx, manifestFiles)
	assert.NotNil(t, err)

	// Scenario 2a: manifest path leads to a directory of manifestFiles - expect success
	manifestFiles, err = getManifestFiles(manifestPathDirectorySuccess)
	assert.Nil(t, err)
	v, err := safeguards.GetManifestViolations(ctx, manifestFiles)
	assert.Nil(t, err)
	assert.Equal(t, len(v), 0)

	// Scenario 2b: manifest path leads to a directory of manifestFiles - expect failure
	manifestFiles, err = getManifestFiles(manifestPathDirectoryError)
	assert.Nil(t, err)
	v, err = safeguards.GetManifestViolations(ctx, manifestFiles)
	assert.Nil(t, err)
	assert.Greater(t, len(v), 0)

	// Scenario 3a: manifest path leads to one manifest file - expect success
	manifestFiles = []string{}
	manifestFiles = append(manifestFiles, manifestPathFileSuccess)
	v, err = safeguards.GetManifestViolations(ctx, manifestFiles)
	assert.Nil(t, err)
	assert.Equal(t, len(v), 0)

	// Scenario 3b: manifest path leads to one manifest file - expect failure
	manifestFiles = []string{}
	manifestFiles = append(manifestFiles, manifestPathFileError)
	v, err = safeguards.GetManifestViolations(ctx, manifestFiles)
	assert.Nil(t, err)
	assert.Greater(t, len(v), 0)
}
