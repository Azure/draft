package templatetests

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestDeploymentManifestTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid manifest deployment",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
			Version:         "0.0.1",
			Dest:            "./validation/.././",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
				"ENVVARS":        `{"key1":"value1","key2":"value2"}`,
			},
		},
		{
			Name:            "valid manifest deployment with filename override",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
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
				"ENVVARS":        `{"key1":"value1","key2":"value2"}`,
			},
			FileNameOverride: map[string]string{
				"deployment.yaml": "deployment-override.yaml",
			},
			UseBaseFixtureWithFileNameOverride: true,
		},
		{
			Name:            "insufficient variables for manifest deployment",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap:          map[string]string{},
			ExpectedErr:     fmt.Errorf("create workflow files: variable APPNAME has no default value"),
		},
		{
			Name:            "manifest deployment vars with err from validators",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
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
			Validators: map[string]func(string) error{
				"kubernetesResourceName": AlwaysFailingValidator,
			},
			ExpectedErr: fmt.Errorf("this is a failing validator"),
		},
		{
			Name:            "manifest deployment vars with err from transformers",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
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
			Transformers: map[string]func(string) (any, error){
				"kubernetesResourceName": AlwaysFailingTransformer,
			},
			ExpectedErr: fmt.Errorf("this is a failing transformer"),
		},
		{
			Name:            "manifest deployment vars with err from label validator",
			TemplateName:    "deployment-manifests",
			FixturesBaseDir: "../../fixtures/deployments/manifest",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"APPNAME":        "*myTestApp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
			Validators: map[string]func(string) error{
				"kubernetesResourceName": K8sLabelValidator,
			},
			ExpectedErr: fmt.Errorf("invalid label: *myTestApp"),
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
