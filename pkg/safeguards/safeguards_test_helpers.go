package safeguards

import (
	"context"
	"testing"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/stretchr/testify/assert"
)

func validateTestManifestsFail(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		errManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		// error case - should throw error
		violations, err := getObjectViolations(ctx, c, errManifests)
		assert.Nil(t, err)
		assert.Greater(t, len(violations), 0)
	}
}

func validateTestManifestsSuccess(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		successManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		// success case - should not throw error
		violations, err := getObjectViolations(ctx, c, successManifests)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(violations))
	}
}
