package cmd

import (
	"context"
	"github.com/Azure/draft/pkg/safeguards"
	"os"
	"path"
	"testing"

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
	testWd, _ := os.Getwd()
	manifestPathEmpty := ""
	manifestPathDirectory := path.Join(testWd, "../pkg/safeguards/tests/container-allowed-images")
	manifestPathFile := path.Join(testWd, "../pkg/safeguards/tests/container-allowed-images/CAI-success-manifest.yaml")
	var manifests []string

	// Scenario 1: empty manifest path should error
	manifests = append(manifests, manifestPathEmpty)
	err := safeguards.ValidateManifests(ctx, manifests)
	assert.NotNil(t, err)

	// Scenario 2: manifest path leads to a directory of manifests
	manifests, err = getManifests(manifestPathDirectory)
	assert.Nil(t, err)
	err = safeguards.ValidateManifests(ctx, manifests)
	assert.Nil(t, err)

	// Scenario 3: manifest path leads to one manifest file
	manifests = append(manifests, manifestPathFile)
	err = safeguards.ValidateManifests(ctx, manifests)
	assert.Nil(t, err)
}
