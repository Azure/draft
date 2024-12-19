package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestManifestsPDBTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid pdb manifest",
			TemplateName:    "podDisruptionBudget-manifests",
			FixturesBaseDir: "../../fixtures/manifests/pdb",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME": "test-app",
				"PARTOF":  "test-app-project",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
