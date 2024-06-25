package safeguards

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// Given a Helm chart directory or file, render all templates and write them to a temporary directory
func RenderHelmChart(isFile bool, mainChartPath, tempDir string) error {
	if isFile { //get the directory that the chart lives in
		mainChartPath = filepath.Dir(mainChartPath)
	}
	loadedCharts := make(map[string]*chart.Chart) // map of chart path to chart object

	mainChart, err := loader.Load(mainChartPath)
	if err != nil {
		return fmt.Errorf("failed to load main chart: %s", err)
	}
	loadedCharts[mainChartPath] = mainChart

	// Load subcharts and dependencies
	for _, dep := range mainChart.Metadata.Dependencies {
		// Resolve the chart path based on the main chart's directory
		chartPath := filepath.Join(mainChartPath, dep.Repository[len("file://"):])
		chartPath = filepath.Clean(chartPath)

		subChart, err := loader.Load(chartPath)
		if err != nil {
			return fmt.Errorf("failed to load chart: %s", err)
		}
		loadedCharts[chartPath] = subChart
	}

	for chartPath, chart := range loadedCharts {
		valuesPath := filepath.Join(chartPath, "values.yaml") // Enforce that values.yaml must be at same level as Chart.yaml
		mergedValues, err := getValues(chart, valuesPath)
		if err != nil {
			return fmt.Errorf("failed to load values: %s", err)
		}
		e := engine.Engine{Strict: true}
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
	}

	return nil
}

func getValues(chart *chart.Chart, valuesPath string) (chartutil.Values, error) {
	// Load values file
	valuesFile, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %s", err)
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesFile, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %s", err)
	}

	mergedValues, err := getReleaseOptions(chart, vals)
	return mergedValues, err
}

func getReleaseOptions(chart *chart.Chart, vals map[string]interface{}) (chartutil.Values, error) {
	// Extract release options from values
	releaseName, ok := vals["releaseName"].(string) //TODO: What do we want to do if a releaseName and namespace is not specified in the values.yaml?
	if !ok || releaseName == "" {
		return nil, fmt.Errorf("releaseName not found or empty in values.yaml")
	}

	releaseNamespace, ok := vals["releaseNamespace"].(string)
	if !ok || releaseNamespace == "" {
		return nil, fmt.Errorf("releaseNamespace not found or empty in values.yaml")
	}

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Namespace: releaseNamespace,
	}

	// Combine chart values with release options
	config := chartutil.Values(vals)
	mergedValues, err := chartutil.ToRenderValues(chart, config, options, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to merge values: %s", err)
	}

	return mergedValues, nil
}
