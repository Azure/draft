package handlers

import (
	"fmt"
	"testing"

	"github.com/Azure/draft/pkg/fixtures"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

func TestManifestDeploymentValidation(t *testing.T) {
	tests := []struct {
		name            string
		templateName    string
		fixturesBaseDir string
		version         string
		dest            string
		templateWriter  *writers.FileMapWriter
		varMap          map[string]string
	}{
		{
			name:            "valid manifest deployment",
			templateName:    "deployment-manifest",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := GetTemplate(tt.templateName, tt.version, tt.dest, tt.templateWriter)
			assert.Nil(t, err)
			assert.NotNil(t, template)

			for k, v := range tt.varMap {
				template.Config.SetVariable(k, v)
			}

			err = template.Generate()
			assert.Nil(t, err)

			for k, v := range tt.templateWriter.FileMap {
				err = fixtures.ValidateContentAgainstFixture(v, fmt.Sprintf("%s/%s", tt.fixturesBaseDir, k))
				assert.Nil(t, err)
			}
		})
	}
}
