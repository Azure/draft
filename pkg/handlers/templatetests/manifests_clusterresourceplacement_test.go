package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestManifestsClusterResourcePlacementTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid clusterresourceplacement manifest with PickAll",
			TemplateName:    "kubefleet-clusterresourceplacement",
			FixturesBaseDir: "../../fixtures/manifests/clusterresourceplacement/pickall",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"CRP_NAME":                "demo-crp",
				"RESOURCE_SELECTOR_NAME":  "fmad-demo", 
				"PLACEMENT_TYPE":          "PickAll",
				"PARTOF":                  "test-app-project",
			},
		},
		{
			Name:            "valid clusterresourceplacement manifest with PickFixed",
			TemplateName:    "kubefleet-clusterresourceplacement", 
			FixturesBaseDir: "../../fixtures/manifests/clusterresourceplacement/pickfixed",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"CRP_NAME":                "fmad-demo-crp",
				"RESOURCE_SELECTOR_NAME":  "fmad-demo",
				"PLACEMENT_TYPE":          "PickFixed",
				"CLUSTER_NAME_1":          "cluster-name-01",
				"CLUSTER_NAME_2":          "cluster-name-02",
				"PARTOF":                  "test-app-project",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}