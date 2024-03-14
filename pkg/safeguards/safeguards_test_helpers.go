package safeguards

import (
	"context"
	"testing"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/stretchr/testify/assert"
)

func validateTestManifests_Error(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		errManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		// error case - should throw error
		violations, err := validateManifests(ctx, c, errManifests)
		assert.NotNil(t, err)
		assert.NotNil(t, violations)
	}
}

func validateTestManifests_Success(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		successManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		// success case - should not throw error
		violations, err := validateManifests(ctx, c, successManifests)
		assert.Nil(t, err)
		assert.Nil(t, violations)
	}
}
