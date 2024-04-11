package cmd

import "github.com/Azure/draft/pkg/safeguards"

func countTestViolations(violations []safeguards.ManifestViolation) int {
	numViolations := 0
	for _, v := range violations {
		numViolations += len(v.ObjectViolations)
	}

	return numViolations
}
