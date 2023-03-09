package example

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestWriteDockerfile(t *testing.T) {
	templateWriter := writers.FileMapWriter{}
	outputPath := "test/path"

	testCases := []struct {
		name               string
		inputVariables     map[string]string
		generationLanguage string
		expectError        bool
	}{

		{
			name: "Test Valid Go Dockerfile Generation",
			inputVariables: map[string]string{
				"PORT":    "8080",
				"VERSION": "1.20",
			},
			generationLanguage: "go",
			expectError:        false,
		},
		{
			name: "Test Invalid Go Dockerfile Generation",
			inputVariables: map[string]string{
				"PORT": "8080",
			},
			generationLanguage: "go",
			expectError:        true,
		},
		{
			name: "Test Invalid GenerationLanguage",
			inputVariables: map[string]string{
				"PORT":    "8080",
				"VERSION": "1.20",
			},
			generationLanguage: "invalid",
			expectError:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := WriteDockerfile(&templateWriter, outputPath, tc.inputVariables, tc.generationLanguage)
			errored := err != nil
			if err != nil {
				fmt.Printf("WriteDockerfile failed: %e\n", err)
			}
			if errored != tc.expectError {
				t.Errorf("WriteDockerfile failed: expected error %t, got %t", tc.expectError, errored)
				t.Fail()
			}
		})
	}
}

func TestWriteDockerfileExample(t *testing.T) {
	err := WriteDockerfileExample()
	if err != nil {
		t.Errorf("WriteDockerfileExample failed: %e", err)
	}
}
