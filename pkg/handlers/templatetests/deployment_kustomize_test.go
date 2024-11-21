package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestDeploymentKustomizeTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid kustomize deployment",
			TemplateName:    "deployment-kustomize",
			FixturesBaseDir: "../../fixtures/deployments/kustomize",
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
