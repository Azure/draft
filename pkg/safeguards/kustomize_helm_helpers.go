package safeguards

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// Given a valid helm chart, renders to ManifestFile and returns the object
func RenderHelmChart(chartPath, valuesPath, tempDir string) error {
	// Load Helm chart
	chart, err := loader.Load(chartPath)
	if err != nil {
		return fmt.Errorf("failed to load chart: %s", err)
	}

	// Load values file
	valuesFile, err := os.ReadFile(valuesPath)
	if err != nil {
		return fmt.Errorf("failed to read values file: %s", err)
	}

	// Parse values.yaml
	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesFile, &vals); err != nil {
		return fmt.Errorf("failed to parse values.yaml: %s", err)
	}

	// Extract release options from values
	releaseName, ok := vals["releaseName"].(string)
	if !ok || releaseName == "" {
		log.Fatalf("releaseName not found or empty in values.yaml")
	}

	releaseNamespace, ok := vals["releaseNamespace"].(string)
	if !ok || releaseNamespace == "" {
		log.Fatalf("releaseNamespace not found or empty in values.yaml")
	}

	// Remove release options from vals map to avoid conflicts
	delete(vals, "releaseName")
	delete(vals, "releaseNamespace")

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: releaseNamespace,
		IsInstall: true,
		IsUpgrade: false,
		Revision:  1,
	}

	// Combine chart values with defaults and release options
	config := chartutil.Values(vals)
	mergedValues, err := chartutil.ToRenderValues(chart, config, options, nil)
	if err != nil {
		return fmt.Errorf("failed to merge values: %s", err)
	}

	e := engine.Engine{Strict: true}
	// Render the templates
	renderedFiles, err := e.Render(chart, mergedValues)
	if err != nil {
		return fmt.Errorf("failed to render chart: %s", err)
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	fmt.Println("Before printing out rendered files ")
	// Write each rendered file to the output directory with the same name as in templates/
	for filePath, content := range renderedFiles {
		fmt.Printf("Generated manifest file: %s\n", tempDir)
		fmt.Println(content)
		fmt.Println("========================================")
		outputFilePath := filepath.Join(tempDir, filepath.Base(filePath))
		if err := os.WriteFile(outputFilePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write manifest file: %s", err)
		}

	}
	return nil
}
