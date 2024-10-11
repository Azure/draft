package handlers

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/fixtures"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

func TestDeepCopy(t *testing.T) {
	// This will fail on adding a new field to the undelying structs that arent handled in DeepCopy
	testTemplate, err := GetTemplate("deployment-manifests", "0.0.1", ".", &writers.FileMapWriter{})
	assert.Nil(t, err)

	deepCopy := testTemplate.DeepCopy()

	assert.True(t, reflect.DeepEqual(deepCopy, testTemplate))
}

func TestTemplateHandlerValidation(t *testing.T) {
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
		{
			name:            "valid clojure dockerfile",
			templateName:    "dockerfile-clojure",
			fixturesBaseDir: "../fixtures/dockerfiles/clojure",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "19-jdk-alpine",
			},
		},
		{
			name:            "valid csharp dockerfile",
			templateName:    "dockerfile-csharp",
			fixturesBaseDir: "../fixtures/dockerfiles/csharp",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "6.0",
			},
		},
		{
			name:            "valid erlang dockerfile",
			templateName:    "dockerfile-erlang",
			fixturesBaseDir: "../fixtures/dockerfiles/erlang",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "27.0-alpine",
				"VERSION":      "3.17",
			},
		},
		{
			name:            "valid go dockerfile",
			templateName:    "dockerfile-go",
			fixturesBaseDir: "../fixtures/dockerfiles/go",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.23",
			},
		},
		{
			name:            "valid gomodule dockerfile",
			templateName:    "dockerfile-gomodule",
			fixturesBaseDir: "../fixtures/dockerfiles/gomodule",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.23",
			},
		},
		{
			name:            "valid gradle dockerfile",
			templateName:    "dockerfile-gradle",
			fixturesBaseDir: "../fixtures/dockerfiles/gradle",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "jdk21",
				"VERSION":      "21-jre",
			},
		},
		{
			name:            "valid gradlew dockerfile",
			templateName:    "dockerfile-gradlew",
			fixturesBaseDir: "../fixtures/dockerfiles/gradlew",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "jdk21",
				"VERSION":      "21-jre",
			},
		},
		{
			name:            "valid java dockerfile",
			templateName:    "dockerfile-java",
			fixturesBaseDir: "../fixtures/dockerfiles/java",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "3 (jdk-21)",
				"VERSION":      "21-jre",
			},
		},
		{
			name:            "valid javascript dockerfile",
			templateName:    "dockerfile-javascript",
			fixturesBaseDir: "../fixtures/dockerfiles/javascript",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "14.15.4",
			},
		},
		{
			name:            "valid php dockerfile",
			templateName:    "dockerfile-php",
			fixturesBaseDir: "../fixtures/dockerfiles/php",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":         "80",
				"BUILDVERSION": "1",
				"VERSION":      "7.1-apache",
			},
		},
		{
			name:            "valid python dockerfile",
			templateName:    "dockerfile-python",
			fixturesBaseDir: "../fixtures/dockerfiles/python",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":       "80",
				"ENTRYPOINT": "app.py",
				"VERSION":    "3.9",
			},
		},
		{
			name:            "valid ruby dockerfile",
			templateName:    "dockerfile-ruby",
			fixturesBaseDir: "../fixtures/dockerfiles/ruby",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "3.1.2",
			},
		},
		{
			name:            "valid rust dockerfile",
			templateName:    "dockerfile-rust",
			fixturesBaseDir: "../fixtures/dockerfiles/rust",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "1.70.0",
			},
		},
		{
			name:            "valid swift dockerfile",
			templateName:    "dockerfile-swift",
			fixturesBaseDir: "../fixtures/dockerfiles/swift",
			version:         "0.0.1",
			dest:            ".",
			templateWriter:  &writers.FileMapWriter{},
			varMap: map[string]string{
				"PORT":    "80",
				"VERSION": "5.5",
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
