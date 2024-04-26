package cmd

import "github.com/Azure/draft/pkg/safeguards"

func countTestViolations(results []safeguards.ManifestResult) int {
	numViolations := 0
	for _, r := range results {
		numViolations += len(r.ObjectViolations)
	}

	return numViolations
}
