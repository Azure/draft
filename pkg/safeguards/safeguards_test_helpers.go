package safeguards

import (
	"context"
	"embed"
	"testing"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/stretchr/testify/assert"
)

//go:embed tests
var testFS embed.FS

func validateTestManifests_Error(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		errManifests, err := testFc.ReadManifests(testFS, path)
		assert.Nil(t, err)

		// error case - should throw error
		err = validateManifests(ctx, c, errManifests)
		assert.NotNil(t, err)
	}
}

func validateTestManifests_Success(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		successManifests, err := testFc.ReadManifests(testFS, path)
		assert.Nil(t, err)

		// success case - should not throw error
		err = validateManifests(ctx, c, successManifests)
		assert.Nil(t, err)
	}
}
