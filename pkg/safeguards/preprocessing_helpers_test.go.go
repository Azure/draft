package safeguards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chartutil"
)

type testVars struct {
	chartYaml    string
	valuesYaml   string
	templates    map[string]string
	invalidChart string
}

func setup() testVars {
	validChartYaml := `apiVersion: v2
name: test-chart
version: 0.1.0
`

	invalidChartYaml := `apiVersion: v2
version: 0.1.0
`
	// Create a minimal values.yaml
	valuesYaml := `replicaCount: 1
`

	templates := make(map[string]string)
	templates["pod.yaml"] = `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx
`
	return testVars{chartYaml: validChartYaml, valuesYaml: valuesYaml, templates: templates, invalidChart: invalidChartYaml}
}
func TestValidateHelmChartSuccess(t *testing.T) {
	t.Parallel()
	v := setup()
	// Create a temporary directory for the test chart
	tmpDir, err := os.MkdirTemp("", "test-helm-chart")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal Chart.yaml

	chartYamlPath := filepath.Join(tmpDir, chartutil.ChartfileName)
	if err := os.WriteFile(chartYamlPath, []byte(v.chartYaml), 0644); err != nil {
		t.Fatalf("Failed to write Chart.yaml for testing: %s", err)
	}

	valuesYamlPath := filepath.Join(tmpDir, chartutil.ValuesfileName)
	if err := os.WriteFile(valuesYamlPath, []byte(v.valuesYaml), 0644); err != nil {
		t.Fatalf("Failed to write values.yaml for testing: %s", err)
	}

	// Create a templates directory and a minimal template
	templatesDir := filepath.Join(tmpDir, chartutil.TemplatesDir)
	if err := os.Mkdir(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory for testing: %s", err)
	}

	templatePath := filepath.Join(templatesDir, "pod.yaml")
	if err := os.WriteFile(templatePath, []byte(v.templates["pod.yaml"]), 0644); err != nil {
		t.Fatalf("Failed to write pod.yaml template for testing: %s", err)
	}

	isHelm, err := isHelmChart(tmpDir)
	assert.True(t, isHelm)
	assert.Nil(t, err)
}

func TestInvalidHelmChart(t *testing.T) {
	t.Parallel()
	v := setup()

	// Create a temporary directory for the test chart
	tmpDir, err := os.MkdirTemp("", "test-helm-chart")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	chartYamlPath := filepath.Join(tmpDir, chartutil.ChartfileName)
	if err := os.WriteFile(chartYamlPath, []byte(v.invalidChart), 0644); err != nil {
		t.Fatalf("Failed to write Chart.yaml for testing: %s", err)
	}

	// Create a minimal values.yaml
	valuesYaml := `replicaCount: 1
`
	valuesYamlPath := filepath.Join(tmpDir, chartutil.ValuesfileName)
	if err := os.WriteFile(valuesYamlPath, []byte(valuesYaml), 0644); err != nil {
		t.Fatalf("Failed to write values.yaml for testing: %s", err)
	}

	// Create a templates directory and a minimal template
	templatesDir := filepath.Join(tmpDir, chartutil.TemplatesDir)
	if err := os.Mkdir(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory for testing: %s", err)
	}
	template := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx
`
	templatePath := filepath.Join(templatesDir, "pod.yaml")
	if err := os.WriteFile(templatePath, []byte(template), 0644); err != nil {
		t.Fatalf("Failed to write pod.yaml template for testing: %s", err)
	}

	isHelm, err := isHelmChart(tmpDir)
	assert.False(t, isHelm)
	assert.NotNil(t, err)
	assert.Contains(t, err, "invalid chart (Chart.yaml): name must not be empty")
}

// func TestGetDirectoryType(t *testing.T) {
// 	// Create a temporary directory for Helm chart
// 	helmDir, err := ioutil.TempDir("", "helm_chart")
// 	if err != nil {
// 		t.Fatalf("Failed to create temporary directory: %s", err)
// 	}
// 	defer os.RemoveAll(helmDir)

// 	if err := ioutil.WriteFile(filepath.Join(helmDir, "Chart.yaml"), []byte("sample chart"), 0644); err != nil {
// 		t.Fatalf("Failed to create Chart.yaml: %s", err)
// 	}
// 	if err := ioutil.WriteFile(filepath.Join(helmDir, "Values.yaml"), []byte("sample values"), 0644); err != nil {
// 		t.Fatalf("Failed to create Values.yaml: %s", err)
// 	}

// 	dirType := GetDirectoryType(helmDir)
// 	if err != nil {
// 		t.Errorf("GetDirectoryType returned an error for Helm chart directory: %s", err)
// 	}
// 	if dirType != Helm {
// 		t.Errorf("Expected directory type Helm, got %s", dirType)
// 	}

// }

// func TestGetDirectoryType_Unknown(t *testing.T) {
// 	// Test with invalid dir - neither a Helm chart nor a Kustomize configuration
// 	unknownDir, err := ioutil.TempDir("", "unknown_directory")
// 	if err != nil {
// 		t.Fatalf("Failed to create temporary directory: %s", err)
// 	}
// 	defer os.RemoveAll(unknownDir)

// 	if err := ioutil.WriteFile(filepath.Join(helmDir, "Chart.yaml"), []byte("sample chart"), 0644); err != nil {
// 		t.Fatalf("Failed to create Chart.yaml: %s", err)
// 	}
// 	if err := ioutil.WriteFile(filepath.Join(helmDir, "values.yaml"), []byte("sample values"), 0644); err != nil {
// 		t.Fatalf("Failed to create values.yaml: %s", err)
// 	}

// 	dirType, err = GetDirectoryType(unknownDir)
// 	if err != nil {
// 		t.Errorf("GetDirectoryType returned an error for unknown directory: %s", err)
// 	}
// 	if dirType != Unknown {
// 		t.Errorf("Expected directory type Unknown, got %s", dirType)
// 	}
// }
