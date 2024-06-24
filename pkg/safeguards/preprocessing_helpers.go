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

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: releaseNamespace,
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

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	// Write each rendered file to the output directory with the same name as in templates/
	for filePath, content := range renderedFiles {
		outputFilePath := filepath.Join(tempDir, filepath.Base(filePath))
		if err := os.WriteFile(outputFilePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write manifest file: %s", err)
		}

	}
	return nil
}
