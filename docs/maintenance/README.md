## How to Release

To make a new release and publish please follow the following steps.

1. Create a branch `publish-x.y.z`
2. Add a section to `CHANGELOG.md` with the header `## [x.y.z]` (N.B: make sure to write the new version in square brackets as the `changelog-reader` action only works if the `CHANGELOG.md` file follows the [Keep a Changelog standard](https://github.com/olivierlacan/keep-a-changelog))
3. Create a new PR, get approval and merge
4. Run the `Draft Release & Publish` workflow manually from the GH Actions tab

Note: `DRAFT_VERSION` can be written as on step env var. For more reference: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable 