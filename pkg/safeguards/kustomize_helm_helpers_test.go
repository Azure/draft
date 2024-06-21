package safeguards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	chartPath   = "testdata/my-web-app"             // Path to the test chart directory
	valuesPath  = "testdata/my-web-app/values.yaml" // Path to the test values.yaml file
	testTempDir = "testdata/output/manifests"       // Path to the output directory
)

type testVars struct {
	validChartYaml         string
	validValuesYaml        string
	validDeploymentYaml    string
	validServiceYaml       string
	validIngressYaml       string
	invalidChartYaml       string
	invalidValuesYaml      string
	invalidTemplateYamls   map[string]string
	expectedValidManifests map[string]string
}

/*
* Invalid values.yaml
* Invalid chart.yaml -- what is the bare minimum needed for a chart.yaml? What fields, if included, would break this function?
* short template files, long template files.
* One template file, multiple template files. One valid, others invalid
 */
func setup(t *testing.T) testVars {
	// Ensure the output directory is empty before running the test
	if err := os.RemoveAll(testTempDir); err != nil {
		t.Fatalf("Failed to clean output directory: %s", err)
	}

	// Create the templates directory
	err := os.MkdirAll(filepath.Join(chartPath, "templates"), 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %s", err)
	}

	// Create Chart.yaml
	chartYAML := `
apiVersion: v2
name: my-web-app
description: A Helm chart for Kubernetes
version: 0.1.0
`

	// Create values.yaml
	valuesYAML := `
replicaCount: 1
image:
  repository: nginx
  tag: stable
service:
  type: ClusterIP
  port: 80
ingress:
  enabled: true
  hostname: example.com
releaseName: test-release
releaseNamespace: test-namespace
`

	// Create templates/deployment.yaml
	deploymentYAML := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-deployment
  namespace: {{ .Release.Namespace }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: my-web-app
  template:
    metadata:
      labels:
        app: my-web-app
    spec:
      containers:
        - name: nginx
          image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
`

	// Create templates/service.yaml
	serviceYAML := `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-service
  namespace: {{ .Release.Namespace }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
  selector:
    app: my-web-app
`

	// Create templates/ingress.yaml
	ingressYAML := `
{{- if .Values.ingress.enabled }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ .Release.Name }}-ingress
  namespace: {{ .Release.Namespace }}
spec:
  rules:
    - host: {{ .Values.ingress.hostname }}
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ .Release.Name }}-service
                port:
                  number: {{ .Values.service.port }}
{{- end }}
`

	// Unspecified name field
	invalidChartYaml := `
description: A Helm chart for Kubernetes
version: 0.1.0
`
	//mising service info
	invalidValuesYaml := `
replicaCount: 1
image:
  repository: nginx
  tag: stable
ingress:
  enabled: true
  hostname: example.com
releaseName: test-release
releaseNamespace: test-namespace
`

	validTemplateYamls := make(map[string]string)
	validTemplateYamls["deployment.yaml"] = deploymentYAML
	validTemplateYamls["service.yaml"] = serviceYAML
	validTemplateYamls["ingress.yaml"] = ingressYAML

	expectedValidManifests := make(map[string]string)
	validTemplateYamls["deployment.yaml"] = getExpectedDeploymentYAML()
	validTemplateYamls["service.yaml"] = getExpectedServiceYAML()
	validTemplateYamls["ingress.yaml"] = getExpectedIngressYAML()

	//invalidDeploymentYaml := strings.Replace(deploymentYAML, "{{ .Values.image.repository }}", "{{ .Values.invalidField }}", 1)
	// invalidServiceYaml := strings.Replace(serviceYAML, "kind: Service\n", "", 1)
	// invalidIngressYaml := strings.Replace(ingressYAML, "kind: Ingress\n", "", 1)

	invalidDeploymentYaml := `
	apiVersion: apps/v1
	kind: Deployment
	metadata:
	  name: {{ .Release.Name }}-deployment
	  namespace: {{ .Release.Namespace }}
	spec:
	  replicas: {{ .Values.replicaCount }}
	  selector:
	    matchLabels:
	      app: my-web-app
	  template:
	    metadata:
	      labels:
	        app: my-web-app
	    spec:
	      containers:
	        - name: nginx
	          image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
	`
	invalidDeploymentSyntaxErr := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-deployment
  namespace: {{ .Release.Namespace }}
spec:
  replicas: {{ .Values.replicaCount
  selector:
    matchLabels:
      app: my-web-app
  template:
    metadata:
      labels:
        app: my-web-ap
    spec:
      containers:
        - name: nginx
          image: {{ .Values.image.repository }}:{{ .Values.image.tag
`
	// Create invalid templates
	invalidDeploymentValues := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-deployment
  namespace: {{ .Release.Namespace }}
spec:
  replicas: {{ .Values.nonExistentField }}
  selector:
    matchLabels:
      app: my-web-app
  template:
    metadata:
      labels:
        app: my-web-app
    spec:
      containers:
        - name: nginx
          image: {{ .Values.image.invalidField }}:{{ .Values.image.tag }}
`
	invalidTemplateYamls := make(map[string]string)
	invalidTemplateYamls["deployment.yaml"] = invalidDeploymentYaml
	invalidTemplateYamls["deploymentSyntax.yaml"] = invalidDeploymentSyntaxErr
	invalidTemplateYamls["deploymentValues.yaml"] = invalidDeploymentValues
	// invalidTemplateYamls["service.yaml"] = se
	// invalidTemplateYamls["ingress.yaml"] = strings.Replace(ingressYAML, "kind: Ingress\n", "", 1)

	//validTemplateYamls: map[string]string{"service.yaml": serviceYAML}
	return testVars{validChartYaml: chartYAML, validValuesYaml: valuesYAML, validDeploymentYaml: deploymentYAML, validServiceYaml: serviceYAML, validIngressYaml: ingressYAML, invalidChartYaml: invalidChartYaml, invalidValuesYaml: invalidValuesYaml, invalidTemplateYamls: invalidTemplateYamls, expectedValidManifests: expectedValidManifests}
}

func TestRenderHelmChart_Valid(t *testing.T) {
	v := setup(t)
	// Register cleanup function to remove output directory after the test
	t.Cleanup(func() { cleanupDir(testTempDir) })

	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = os.WriteFile(valuesPath, []byte(v.validValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.validDeploymentYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/service.yaml"), []byte(v.validServiceYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/service.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/ingress.yaml"), []byte(v.validIngressYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/ingress.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, testTempDir)
	assert.Nil(t, err)

	// Check that the output directory exists and contains expected files
	expectedFiles := []string{"deployment.yaml", "service.yaml", "ingress.yaml"}
	for _, fileName := range expectedFiles {
		outputFilePath := filepath.Join(testTempDir, fileName)
		assert.FileExists(t, outputFilePath, "Expected file does not exist: %s", outputFilePath)
	}

	//assert that each file output matches expected yaml after values are filled in
	assert.Equal(t, parseYAML(t, getExpectedDeploymentYAML()), parseYAML(t, readFile(t, filepath.Join(testTempDir, "deployment.yaml"))))
	assert.Equal(t, parseYAML(t, getExpectedServiceYAML()), parseYAML(t, readFile(t, filepath.Join(testTempDir, "service.yaml"))))
	assert.Equal(t, parseYAML(t, getExpectedIngressYAML()), parseYAML(t, readFile(t, filepath.Join(testTempDir, "ingress.yaml"))))
}

/**
* Testing user errors
 */

// Should fail if the chart and values.yaml are invalid
func TestInvalidChartAndValues(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(testTempDir) })

	// Invalid Chart.yaml
	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.invalidChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = RenderHelmChart(chartPath, valuesPath, testTempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to load chart: validation: chart.metadata.name is required")

	//assert.Contains(t, err.Error(), "executing \"service.yaml\" at <.Values.service.type>: map has no entry for key \"service\"")
	// Write valid Chart.yaml
	err = os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	// Write invalid values.yaml
	err = os.WriteFile(filepath.Join(chartPath, "values.yaml"), []byte(v.invalidValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	// Write the service template
	err = os.WriteFile(filepath.Join(chartPath, "templates/service.yaml"), []byte(v.validServiceYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/service.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, testTempDir)

	// Assert that an error occurs
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "<.Values.service.type>: nil pointer evaluating interface {}.type")
}

func TestInvalidTemplate(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(testTempDir) })

	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = os.WriteFile(valuesPath, []byte(v.validValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.invalidTemplateYamls["deploymentValues.yaml"]), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, testTempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to render chart: template: my-web-app/templates/deployment.yaml")
	assert.Contains(t, err.Error(), "map has no entry for key \"nonExistentField\"")

	cleanupDir(testTempDir)
	err = os.WriteFile(filepath.Join(chartPath, "templates/deployment.yaml"), []byte(v.invalidTemplateYamls["deploymentSyntax.yaml"]), 0644)
	if err != nil {
		t.Fatalf("Failed to write templates/deployment.yaml: %s", err)
	}

	// Run the function
	err = RenderHelmChart(chartPath, valuesPath, testTempDir)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse error")
	assert.Contains(t, err.Error(), "function \"selector\" not defined")
}

func TestInvalidTemplateSyntax(t *testing.T) {
	v := setup(t)
	t.Cleanup(func() { cleanupDir(testTempDir) })

	err := os.WriteFile(filepath.Join(chartPath, "Chart.yaml"), []byte(v.validChartYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write Chart.yaml: %s", err)
	}

	err = os.WriteFile(valuesPath, []byte(v.validValuesYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write values.yaml: %s", err)
	}

	// assert.Contains(t, err.Error(), "failed to render chart: template: my-web-app/templates/deployment.yaml")
	// assert.Contains(t, err.Error(), "map has no entry for key \"nonExistentField\"")

}

func cleanupDir(dir string) {
	os.RemoveAll(dir)
}

func getExpectedDeploymentYAML() string {
	return `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-release-deployment
  namespace: test-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-web-app
  template:
    metadata:
      labels:
        app: my-web-app
    spec:
      containers:
        - name: nginx
          image: nginx:stable
`
}

func getExpectedServiceYAML() string {
	return `
apiVersion: v1
kind: Service
metadata:
  name: test-release-service
  namespace: test-namespace
spec:
  type: ClusterIP
  ports:
    - port: 80
  selector:
    app: my-web-app
`
}

func getExpectedIngressYAML() string {
	return `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-release-ingress
  namespace: test-namespace
spec:
  rules:
    - host: example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: test-release-service
                port:
                  number: 80
`
}

func readFile(t *testing.T, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %s", err)
	}
	return string(data)
}

func parseYAML(t *testing.T, content string) map[string]interface{} {
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &result)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %s", err)
	}
	return result
}
