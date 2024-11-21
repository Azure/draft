package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestDeploymentHelmTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid helm deployment",
			TemplateName:    "deployment-helm",
			FixturesBaseDir: "../../fixtures/deployments/helm",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
