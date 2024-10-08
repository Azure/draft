package handlers

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/fixtures"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

func TestManifestDeploymentValidation(t *testing.T) {
	tests := []struct {
		name             string
		templateName     string
		fixturesBaseDir  string
		version          string
		dest             string
		templateWriter   *writers.FileMapWriter
		varMap           map[string]string
		fileNameOverride map[string]string
		expectedErr      error
	}{
		{
			name:            "valid manifest deployment",
			templateName:    "deployment-manifests",
			fixturesBaseDir: "../fixtures/deployments/manifest",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
		},
		{
			name:            "valid helm deployment",
			templateName:    "deployment-helm",
			fixturesBaseDir: "../fixtures/deployments/helm",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
		},
		{
			name:            "valid kustomize deployment",
			templateName:    "deployment-kustomize",
			fixturesBaseDir: "../fixtures/deployments/kustomize",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
		},
		{
			name:            "valid manifest deployment with filename override",
			templateName:    "deployment-manifests",
			fixturesBaseDir: "../fixtures/deployments/manifest",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"APPNAME":        "testapp",
				"NAMESPACE":      "default",
				"PORT":           "80",
				"IMAGENAME":      "testimage",
				"IMAGETAG":       "latest",
				"GENERATORLABEL": "draft",
				"SERVICEPORT":    "80",
			},
			fileNameOverride: map[string]string{
				"deployment.yaml": "deployment-override.yaml",
			},
		},
		{
			name:            "insufficient variables for manifest deployment",
			templateName:    "deployment-manifests",
			fixturesBaseDir: "../fixtures/deployments/manifest",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap:          map[string]string{},
			expectedErr:     fmt.Errorf("create workflow files: variable APPNAME has no default value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := GetTemplate(tt.templateName, tt.version, tt.dest, tt.templateWriter)
			assert.Nil(t, err)
			assert.NotNil(t, template)

			for k, v := range tt.varMap {
				template.Config.SetVariable(k, v)
			}

			overrideReverseLookup := make(map[string]string)
			for k, v := range tt.fileNameOverride {
				template.Config.SetFileNameOverride(k, v)
				overrideReverseLookup[v] = k
			}

			err = template.Generate()
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)

			for k, v := range tt.templateWriter.FileMap {
				fileName := k
				if overrideFile, ok := overrideReverseLookup[filepath.Base(k)]; ok {
					fileName = strings.Replace(fileName, filepath.Base(k), overrideFile, 1)
				}

				err = fixtures.ValidateContentAgainstFixture(v, fmt.Sprintf("%s/%s", tt.fixturesBaseDir, fileName))
				assert.Nil(t, err)
			}
		})
	}
}
