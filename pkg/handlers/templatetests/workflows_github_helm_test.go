package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestGitHubWorkflowHelmTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid helm workflow",
			TemplateName:    "github-workflow-helm",
			FixturesBaseDir: "../../fixtures/workflows/github/helm",
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
				"CLUSTERRESOURCETYPE":    "managedClusters",
				"CLUSTERNAME":            "testCluster",
				"KUSTOMIZEPATH":          "./overlays/production",
				"DEPLOYMENTMANIFESTPATH": "./manifests",
				"DOCKERFILE":             "./Dockerfile",
				"BUILDCONTEXTPATH":       "test",
				"CHARTPATH":              "testPath",
				"CHARTOVERRIDEPATH":      "testOverridePath",
				"CHARTOVERRIDES":         "replicas:2",
				"NAMESPACE":              "default",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
