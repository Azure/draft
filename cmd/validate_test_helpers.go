package cmd

import (
	types "github.com/Azure/draft/pkg/safeguards/types"
)

func countTestViolations(results []types.ManifestResult) int {
	numViolations := 0
	for _, r := range results {
		numViolations += len(r.ObjectViolations)
	}

	return numViolations
}
