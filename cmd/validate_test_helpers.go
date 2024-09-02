package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bfoley13/draft/pkg/safeguards"
)

var tempDir, _ = filepath.Abs("./testdata")

func countTestViolations(results []safeguards.ManifestResult) int {
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
