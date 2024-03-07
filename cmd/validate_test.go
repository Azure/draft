package cmd

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/Azure/draft/pkg/safeguards"
	"github.com/stretchr/testify/assert"
)

var wd, _ = os.Getwd()
var testFS = os.DirFS(path.Join(wd, "../pkg/safeguards/tests"))
var emptyFS = os.DirFS("")

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
	manifestPathDirectorySuccess := "all/success"
	manifestPathDirectoryError := "all/error"
	manifestPathFileSuccess := "all/success/all-success-manifest.yaml"
	manifestPathFileError := "all/error/all-error-manifest.yaml"
	var manifests []string

	// Scenario 1: empty manifest path should error
	manifests = append(manifests, manifestPathEmpty)
	err := safeguards.ValidateManifests(ctx, emptyFS, manifests)
	assert.NotNil(t, err)

	// Scenario 2a: manifest path leads to a directory of manifests - expect success
	manifests, err = getManifests(testFS, manifestPathDirectorySuccess)
	assert.Nil(t, err)
	err = safeguards.ValidateManifests(ctx, testFS, manifests)
	assert.Nil(t, err)

	// Scenario 2b: manifest path leads to a directory of manifests - expect failure
	manifests, err = getManifests(testFS, manifestPathDirectoryError)
	assert.Nil(t, err)
	err = safeguards.ValidateManifests(ctx, testFS, manifests)
	assert.NotNil(t, err)

	// Scenario 3a: manifest path leads to one manifest file - expect success
	manifests = []string{}
	manifests = append(manifests, manifestPathFileSuccess)
	err = safeguards.ValidateManifests(ctx, testFS, manifests)
	assert.Nil(t, err)

	// Scenario 3b: manifest path leads to one manifest file - expect failure
	manifests = []string{}
	manifests = append(manifests, manifestPathFileError)
	err = safeguards.ValidateManifests(ctx, testFS, manifests)
	assert.NotNil(t, err)
}
