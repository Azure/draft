package example

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestWriteDeploymentFiles(t *testing.T) {
	filewriter := writers.FileMapWriter{}
	outputPath := "test/path"

	testCases := []struct {
		name           string
		deployConfig   *config.DraftConfig
		deploymentType string
		expectError    bool
	}{
		{
			name: "Test Valid Manifests Deployment Generation",
			deployConfig: &config.DraftConfig{
				Variables: []*config.BuilderVar{
					{
						Name: "PORT",
						Default: config.BuilderVarDefault{
							Value: "80",
						},
						Description: "the port exposed in the application",
						Value:       "8080",
					},
					{
						Name:        "APPNAME",
						Description: "the name of the application",
						Value:       "testapp",
					},
					{
						Name: "SERVICEPORT",
						Default: config.BuilderVarDefault{
							ReferenceVar: "PORT",
						},
						Description: "the port the service uses to make the application accessible from outside the cluster",
						Value:       "8080",
					},
					{
						Name: "NAMESPACE",
						Default: config.BuilderVarDefault{
							Value: "default",
						},
						Description: "the name of the image to use in the deployment",
						Value:       "testnamespace",
					},
					{
						Name: "IMAGENAME",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "the name of the image to use in the deployment",
						},
						Description: "the name of the image to use in the deployment",
						Value:       "testimage",
					},
					{
						Name: "IMAGETAG",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "latest",
						},
						Description: "the tag of the image to use in the deployment",
						Value:       "latest",
					},
					{
						Name: "GENERATORLABEL",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "draft",
						},
						Description: "the label to use to identify the deployment as generated by draft",
					},
				},
			},
			deploymentType: "manifests",
			expectError:    false,
		},
		{
			name: "Test Invalid Manifests Deployment Generation",
			deployConfig: &config.DraftConfig{
				Variables: []*config.BuilderVar{
					{
						Name: "PORT",
						Default: config.BuilderVarDefault{
							Value: "80",
						},
						Description: "the port exposed in the application",
						Value:       "8080",
					},
					{
						Name:        "APPNAME",
						Description: "the name of the application",
					},
					{
						Name: "SERVICEPORT",
						Default: config.BuilderVarDefault{
							ReferenceVar: "PORT",
						},
						Description: "the port the service uses to make the application accessible from outside the cluster",
					},
					{
						Name: "NAMESPACE",
						Default: config.BuilderVarDefault{
							Value: "default",
						},
						Description: "the name of the image to use in the deployment",
					},
					{
						Name: "IMAGENAME",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "the name of the image to use in the deployment",
						},
						Description: "the name of the image to use in the deployment",
					},
					{
						Name: "IMAGETAG",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "latest",
						},
						Description: "the tag of the image to use in the deployment",
					},
					{
						Name: "GENERATORLABEL",
						Default: config.BuilderVarDefault{
							IsPromptDisabled: true,
							Value:            "draft",
						},
						Description: "the label to use to identify the deployment as generated by draft",
					},
				},
			},
			deploymentType: "manifests",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := WriteDeploymentFiles(&filewriter, outputPath, tc.deployConfig, tc.deploymentType)
			errored := err != nil
			if err != nil {
				fmt.Printf("WriteDeploymentFiles failed: %e\n", err)
			}
			if errored != tc.expectError {
				t.Errorf("WriteDeploymentFiles failed: expected error %t, got %t", tc.expectError, errored)
				t.Fail()
			}
		})
	}
}

func TestWriteDeploymentFilesExample(t *testing.T) {
	err := WriteDeploymentFilesExample()
	if err != nil {
		t.Errorf("WriteDockerfileExample failed: %e", err)
		t.Fail()
	}
}
