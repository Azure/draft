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

// TestRunValidate tests the run command for `draft validate` for proper returns
func TestRunValidate(t *testing.T) {
	ctx := context.TODO()
	manifestPathEmpty := ""
	manifestPathDirectorySuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success")
	manifestPathDirectoryError, _ := filepath.Abs("../pkg/safeguards/tests/all/error")
	manifestPathFileSuccess, _ := filepath.Abs("../pkg/safeguards/tests/all/success/all-success-manifest.yaml")
	manifestPathFileError, _ := filepath.Abs("../pkg/safeguards/tests/all/error/all-error-manifest.yaml")
	var manifestFiles []string

	// Scenario 1: empty manifest path should error
	manifestFiles = append(manifestFiles, manifestPathEmpty)
	err := safeguards.ValidateManifests(ctx, manifestFiles)
	assert.NotNil(t, err)

	// Scenario 2a: manifest path leads to a directory of manifestFiles - expect success
	manifestFiles, err = getManifestFiles(manifestPathDirectorySuccess)
	assert.Nil(t, err)
	err = safeguards.ValidateManifests(ctx, manifestFiles)
	assert.Nil(t, err)

	// Scenario 2b: manifest path leads to a directory of manifestFiles - expect failure
	manifestFiles, err = getManifestFiles(manifestPathDirectoryError)
	assert.Nil(t, err)
	err = safeguards.ValidateManifests(ctx, manifestFiles)
	assert.NotNil(t, err)

	// Scenario 3a: manifest path leads to one manifest file - expect success
	manifestFiles = []string{}
	manifestFiles = append(manifestFiles, manifestPathFileSuccess)
	err = safeguards.ValidateManifests(ctx, manifestFiles)
	assert.Nil(t, err)

	// Scenario 3b: manifest path leads to one manifest file - expect failure
	manifestFiles = []string{}
	manifestFiles = append(manifestFiles, manifestPathFileError)
	err = safeguards.ValidateManifests(ctx, manifestFiles)
	assert.NotNil(t, err)
}
