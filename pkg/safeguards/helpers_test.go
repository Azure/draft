package safeguards

import (
	"context"
	"testing"

	"github.com/Azure/draft/pkg/safeguards/preprocessing"
	c "github.com/Azure/draft/pkg/safeguards/types"
	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/stretchr/testify/assert"

	"helm.sh/helm/v3/pkg/chartutil"
)

func validateOneTestManifestFail(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc c.FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		errManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		err = loadManifestObjects(ctx, c, errManifests)
		assert.Nil(t, err)

		// error case - should throw error
		violations, err := getObjectViolations(ctx, c, errManifests)
		assert.Nil(t, err)
		assert.Greater(t, len(violations), 0)
	}
}

func validateOneTestManifestSuccess(ctx context.Context, t *testing.T, c *constraintclient.Client, testFc c.FileCrawler, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		successManifests, err := testFc.ReadManifests(path)
		assert.Nil(t, err)

		err = loadManifestObjects(ctx, c, successManifests)
		assert.Nil(t, err)

		// success case - should not throw error
		violations, err := getObjectViolations(ctx, c, successManifests)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(violations))
	}
}

func validateAllTestManifestsFail(ctx context.Context, t *testing.T, testManifestPaths []string) {
	var opt chartutil.ReleaseOptions
	for _, path := range testManifestPaths {
		manifestFiles, err := preprocessing.GetManifestFiles(path, opt)
		assert.Nil(t, err)

		// error case - should throw error
		results, err := GetManifestResults(ctx, manifestFiles)
		for _, mr := range results {
			assert.Nil(t, err)
			assert.Greater(t, mr.ViolationsCount, 0)
		}
	}
}

func validateAllTestManifestsSuccess(ctx context.Context, t *testing.T, testManifestPaths []string) {
	for _, path := range testManifestPaths {
		manifestFiles, err := preprocessing.GetManifestFilesFromDir(path)
		assert.Nil(t, err)

		// success case - should not throw error
		results, err := GetManifestResults(ctx, manifestFiles)
		for _, mr := range results {
			assert.Nil(t, err)
			assert.Equal(t, mr.ViolationsCount, 0)
		}
	}
}
