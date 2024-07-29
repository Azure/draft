package cmd

import (
	"github.com/Azure/draft/pkg/safeguards/types"
)

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
