package cmd

import (
	"os"
	"path/filepath"
	"testing"

	types "github.com/Azure/draft/pkg/safeguards/types"
)

var tempDir, _ = filepath.Abs("./testdata")

const (
	chartPath         = "../pkg/safeguards/tests/testmanifests/validchart"
	kustomizationPath = "../pkg/safeguards/tests/kustomize/overlays/production"
)

func countTestViolations(results []types.ManifestResult) int {
	numViolations := 0
	for _, r := range results {
		numViolations += len(r.ObjectViolations)
	}

	return numViolations
}

func makeTempDir(t *testing.T) {
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("failed to create temporary output directory: %s", err)
	}
}

func cleanupDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to clean directory: %s", err)
	}
}
