package templatetests

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestManifestsHPATemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid hpa manifest",
			TemplateName:    "horizontalPodAutoscaler-manifests",
			FixturesBaseDir: "../../fixtures/manifests/hpa",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME": "test-app",
				"PARTOF":  "test-app-project",
			},
		},
		{
			Name:            "valid hpa manifest with memory utilization",
			TemplateName:    "horizontalPodAutoscaler-manifests",
			FixturesBaseDir: "../../fixtures/manifests/hpa",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME":      "test-app",
				"PARTOF":       "test-app-project",
				"RESOURCETYPE": "memory",
			},
			FileNameOverride: map[string]string{
				"hpa.yaml": "hpa-memory.yaml",
			},
		},
		{
			Name:            "invalid hpa manifest with invalid resource type",
			TemplateName:    "horizontalPodAutoscaler-manifests",
			FixturesBaseDir: "../../fixtures/manifests/hpa",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME":      "test-app",
				"PARTOF":       "test-app-project",
				"RESOURCETYPE": "http",
			},
			ExpectedErr: fmt.Errorf("invalid scaling resource type"),
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
