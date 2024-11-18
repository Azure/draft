package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestGitHubWorkflowKustomizeTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid kustomize workflow",
			TemplateName:    "github-workflow-kustomize",
			FixturesBaseDir: "../../fixtures/workflows/github/kustomize",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"WORKFLOWNAME":           "testWorkflow",
				"BRANCHNAME":             "testBranch",
				"ACRRESOURCEGROUP":       "testAcrRG",
				"AZURECONTAINERREGISTRY": "testAcr",
				"CONTAINERNAME":          "testContainer",
				"CLUSTERRESOURCEGROUP":   "testClusterRG",
				"CLUSTERNAME":            "testCluster",
				"DEPLOYMENTMANIFESTPATH": "./manifests",
				"DOCKERFILE":             "./Dockerfile",
				"BUILDCONTEXTPATH":       "test",
				"NAMESPACE":              "default",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
