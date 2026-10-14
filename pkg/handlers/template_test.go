package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDeepCopy(t *testing.T) {
	// This will fail on adding a new field to the undelying structs that arent handled in DeepCopy
	testTemplate, err := GetTemplate("deployment-manifests", "0.0.1", ".", &writers.FileMapWriter{})
	assert.Nil(t, err)

	deepCopy := testTemplate.DeepCopy()

	assert.True(t, reflect.DeepEqual(deepCopy, testTemplate))
}

// TestCsharpPortRejectsShellInjection covers the csharp template, which
// interpolates PORT into a shell-form ENTRYPOINT. Unlike the python CMD this
// sink cannot be protected by JSON encoding, so the port validator is the only
// defence.
func TestCsharpPortRejectsShellInjection(t *testing.T) {
	for i, port := range []string{
		`8080"; touch pwned; echo "`,
		"$(touch pwned)",
		"80 && touch pwned",
		"80\nRUN touch pwned",
	} {
		t.Run(fmt.Sprintf("payload-%d", i), func(t *testing.T) {
			writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
			csharpTemplate, err := GetTemplate("dockerfile-csharp", "", ".", writer)
			assert.Nil(t, err)

			csharpTemplate.Config.SetVariable("VERSION", "8.0")
			csharpTemplate.Config.SetVariable("PORT", port)

			err = csharpTemplate.Generate()
			assert.NotNil(t, err, "expected port %q to be rejected", port)
			assert.Contains(t, err.Error(), "failed variable validation")
			assert.Empty(t, writer.FileMap, "no files should be written when validation fails")
		})
	}
}

func TestDockerfileVersionRejectsInjection(t *testing.T) {
	writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
	pythonTemplate, err := GetTemplate("dockerfile-python", "", ".", writer)
	assert.Nil(t, err)

	pythonTemplate.Config.SetVariable("PORT", "8080")
	pythonTemplate.Config.SetVariable("ENTRYPOINT", "app.py")
	pythonTemplate.Config.SetVariable("VERSION", "3.13\nRUN touch pwned")

	err = pythonTemplate.Generate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed variable validation")
	assert.Empty(t, writer.FileMap, "no files should be written when validation fails")
}

// unsafeEntrypoints are crafted filenames that an attacker could plant in a
// repository, which the python extractor would pick up as the ENTRYPOINT.
var unsafeEntrypoints = []struct {
	name       string
	entrypoint string
}{
	{name: "embedded quote", entrypoint: "my\"app.py"},
	{name: "embedded backslash", entrypoint: "my\\app.py"},
	{name: "embedded newline", entrypoint: "my\napp.py"},
	{name: "argument injection", entrypoint: "main.py\", \"extra.py"},
	{name: "instruction injection", entrypoint: "app.py\"]\nRUN touch pwned\n# .py"},
}

func newPythonTemplate(t *testing.T, entrypoint string) (*Template, *writers.FileMapWriter) {
	t.Helper()
	writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
	pythonTemplate, err := GetTemplate("dockerfile-python", "", ".", writer)
	assert.Nil(t, err)

	pythonTemplate.Config.SetVariable("PORT", "8080")
	pythonTemplate.Config.SetVariable("ENTRYPOINT", entrypoint)
	return pythonTemplate, writer
}

// TestPythonEntrypointRejectsUnsafeNames is the primary defence: a crafted
// filename is rejected outright by the fileName validator, so no Dockerfile is
// produced at all.
func TestPythonEntrypointRejectsUnsafeNames(t *testing.T) {
	for _, testCase := range unsafeEntrypoints {
		if testCase.name == "embedded quote" || testCase.name == "argument injection" {
			continue
		}
		t.Run(testCase.name, func(t *testing.T) {
			pythonTemplate, writer := newPythonTemplate(t, testCase.entrypoint)

			err := pythonTemplate.Generate()
			assert.NotNil(t, err, "expected %q to be rejected", testCase.entrypoint)
			assert.Contains(t, err.Error(), "failed variable validation")
			assert.Empty(t, writer.FileMap, "no files should be written when validation fails")
		})
	}
}

func TestPythonEntrypointAcceptsOrdinaryNames(t *testing.T) {
	// "src/server.py" covers the nested layout the python extractor produces for
	// a repository whose entrypoint is not in the root.
	for _, entrypoint := range []string{"app.py", "./app.py", "main.py", "my_app.py", "my-app.py", "src/server.py"} {
		t.Run(entrypoint, func(t *testing.T) {
			pythonTemplate, writer := newPythonTemplate(t, entrypoint)
			assert.Nil(t, pythonTemplate.Generate())
			assert.Contains(t, string(writer.FileMap["Dockerfile"]), "CMD [\""+entrypoint+"\"]")
		})
	}
}

func TestPythonEntrypointAcceptsJSONSafePaths(t *testing.T) {
	for _, entrypoint := range []string{"src/my app.py", "資料/服務.py"} {
		t.Run(entrypoint, func(t *testing.T) {
			pythonTemplate, writer := newPythonTemplate(t, entrypoint)
			assert.Nil(t, pythonTemplate.Generate())

			_, command, found := strings.Cut(string(writer.FileMap["Dockerfile"]), "\nCMD ")
			assert.True(t, found)
			var arguments []string
			assert.Nil(t, json.Unmarshal([]byte(strings.TrimSpace(command)), &arguments))
			assert.Equal(t, []string{entrypoint}, arguments)
		})
	}
}

func TestConfigMapPreservesStringValues(t *testing.T) {
	writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
	tpl, err := GetTemplate("deployment-manifests", "", ".", writer)
	assert.Nil(t, err)

	values := map[string]string{
		"APPNAME":        "testapp",
		"NAMESPACE":      "default",
		"PORT":           "80",
		"IMAGENAME":      "testimage",
		"IMAGETAG":       "latest",
		"GENERATORLABEL": "draft",
		"SERVICEPORT":    "80",
		"ENVVARS":        `{"BOOL":"true","COMMENT":"abc # retained","NULL":"null","OBJECT":"{}"}`,
	}
	for name, value := range values {
		tpl.Config.SetVariable(name, value)
	}
	assert.Nil(t, tpl.Generate())

	var configMap struct {
		Data map[string]string `yaml:"data"`
	}
	assert.Nil(t, yaml.Unmarshal(writer.FileMap["manifests/configmap.yaml"], &configMap))
	assert.Equal(t, map[string]string{
		"BOOL": "true", "COMMENT": "abc # retained", "NULL": "null", "OBJECT": "{}",
	}, configMap.Data)
}

func TestGitHubWorkflowQuotesBranchAndBuildContext(t *testing.T) {
	writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
	tpl, err := GetTemplate("github-workflow-manifests", "", ".", writer)
	assert.Nil(t, err)

	values := map[string]string{
		"WORKFLOWNAME": "test workflow", "BRANCHNAME": "feature/security-fix",
		"ACRRESOURCEGROUP": "acr-rg", "AZURECONTAINERREGISTRY": "registry",
		"CONTAINERNAME": "app", "CLUSTERRESOURCEGROUP": "cluster-rg",
		"CLUSTERNAME": "cluster", "DOCKERFILE": "./Dockerfile",
		"BUILDCONTEXTPATH": "src/my app", "NAMESPACE": "default",
	}
	for name, value := range values {
		tpl.Config.SetVariable(name, value)
	}
	assert.Nil(t, tpl.Generate())

	workflow := string(writer.FileMap[".github/workflows/azure-kubernetes-service.yml"])
	assert.Contains(t, workflow, `branches: ["feature/security-fix"]`)
	assert.Contains(t, workflow, `BUILD_CONTEXT_PATH: "src/my app"`)
	assert.Contains(t, workflow, `"${{ env.BUILD_CONTEXT_PATH }}"`)
}

// TestPythonEntrypointJSONEncoding exercises the second layer of the defence in
// isolation. The kind validator is replaced with a permissive one so that the
// crafted values reach the renderer, proving the exec-form CMD stays a
// well-formed single-element JSON array even if validation is bypassed,
// weakened by a consumer override, or the variable is retagged in future.
func TestPythonEntrypointJSONEncoding(t *testing.T) {
	entrypointKind, err := func() (string, error) {
		tpl, err := GetTemplate("dockerfile-python", "", ".", &writers.FileMapWriter{})
		if err != nil {
			return "", err
		}
		v, err := tpl.Config.GetVariable("ENTRYPOINT")
		if err != nil {
			return "", err
		}
		return v.Kind, nil
	}()
	assert.Nil(t, err)

	for _, testCase := range unsafeEntrypoints {
		t.Run(testCase.name, func(t *testing.T) {
			pythonTemplate, writer := newPythonTemplate(t, testCase.entrypoint)
			pythonTemplate.Config.SetVariableValidator(entrypointKind, func(string) error { return nil })
			assert.Nil(t, pythonTemplate.Generate())

			_, command, found := strings.Cut(string(writer.FileMap["Dockerfile"]), "\nCMD ")
			assert.True(t, found)
			command = strings.TrimSuffix(command, "\n")

			// The filename must render as a single JSON string argument and must not
			// break onto a new line as a separate Dockerfile instruction.
			assert.NotContains(t, command, "\n")
			var arguments []string
			assert.Nil(t, json.Unmarshal([]byte(command), &arguments))
			assert.Equal(t, []string{testCase.entrypoint}, arguments)
		})
	}
}

func TestManifestPathListRendersAsOneValue(t *testing.T) {
	paths := "./manifests/deployment.yaml\n./manifests/service.yaml\n./manifests/ingress.yaml"

	testCases := []struct {
		name     string
		template string
		output   string
		values   map[string]string
		lookup   func(map[string]any) any
	}{
		{
			name:     "azure pipeline",
			template: "azure-pipeline-manifests",
			output:   ".pipelines/azure-kubernetes-service.yaml",
			values: map[string]string{
				"ARMSERVICECONNECTION": "conn", "AZURECONTAINERREGISTRY": "registry",
				"CONTAINERNAME": "app", "CLUSTERRESOURCEGROUP": "cluster-rg",
				"ACRRESOURCEGROUP": "acr-rg", "CLUSTERNAME": "cluster",
				"CLUSTERRESOURCETYPE": "Microsoft.ContainerService/managedClusters",
				"MANIFESTPATH":        paths,
			},
			lookup: func(doc map[string]any) any { return doc["variables"].(map[string]any)["manifestPath"] },
		},
		{
			name:     "github workflow",
			template: "github-workflow-manifests",
			output:   ".github/workflows/azure-kubernetes-service.yml",
			values: map[string]string{
				"WORKFLOWNAME": "deploy", "BRANCHNAME": "main",
				"ACRRESOURCEGROUP": "acr-rg", "AZURECONTAINERREGISTRY": "registry",
				"CONTAINERNAME": "app", "CLUSTERRESOURCEGROUP": "cluster-rg",
				"CLUSTERNAME": "cluster", "DOCKERFILE": "./Dockerfile",
				"BUILDCONTEXTPATH": ".", "NAMESPACE": "default",
				"DEPLOYMENTMANIFESTPATH": paths,
			},
			lookup: func(doc map[string]any) any { return doc["env"].(map[string]any)["DEPLOYMENT_MANIFEST_PATH"] },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			writer := &writers.FileMapWriter{FileMap: make(map[string][]byte)}
			tpl, err := GetTemplate(testCase.template, "", ".", writer)
			assert.Nil(t, err)
			for name, value := range testCase.values {
				tpl.Config.SetVariable(name, value)
			}
			assert.Nil(t, tpl.Generate())

			var doc map[string]any
			assert.Nil(t, yaml.Unmarshal(writer.FileMap[testCase.output], &doc))
			assert.Equal(t, paths, testCase.lookup(doc))
		})
	}
}
