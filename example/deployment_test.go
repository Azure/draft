package example

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestWriteDeploymentFiles(t *testing.T) {
	filewriter := writers.FileMapWriter{}
	outputPath := "test/path"

	testCases := []struct {
		name           string
		inputVariables map[string]string
		deploymentType string
		expectError    bool
	}{
		{
			name: "Test Valid Manifests Deployment Generation",
			inputVariables: map[string]string{
				"PORT":        "8080",
				"APPNAME":     "testapp",
				"SERVICEPORT": "8080",
				"NAMESPACE":   "testnamespace",
				"IMAGENAME":   "testimage",
				"IMAGETAG":    "latest",
			},
			deploymentType: "manifests",
			expectError:    false,
		},
		{
			name: "Test Invalid Manifests Deployment Generation",
			inputVariables: map[string]string{
				"PORT": "8080",
			},
			deploymentType: "manifests",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := WriteDeploymentFiles(&filewriter, outputPath, tc.inputVariables, tc.deploymentType)
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
