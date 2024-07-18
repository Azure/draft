package preprocessing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/safeguards"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Returns values from values.yaml and release options specified in values.yaml
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
	releaseName, ok := vals["releaseName"].(string)
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

// IsKustomize checks whether a given path should be treated as a kustomize project
func isKustomize(isDir bool, p string) bool {
	var err error
	if isDir {
		if _, err = os.Stat(filepath.Join(p, "kustomization.yaml")); err == nil {
			return true
		} else if _, err = os.Stat(filepath.Join(p, "kustomization.yml")); err == nil {
			return true
		} else {
			return false
		}
	} else {
		return strings.Contains(p, "kustomization.yaml")
	}
}

// Checks whether a given path is a helm directory or a path to a Helm Chart (contains/is Chart.yaml)
func isHelm(isDir bool, path string) bool {
	var chartPath string
	if isDir {
		chartPath = filepath.Join(path, "Chart.yaml")
	} else {
		chartPath = path
	}

	_, err := os.Stat(chartPath)
	if err == nil && safeguards.IsYAML(chartPath) { // Couldn't find Chart.yaml in the directory
		return true
	}

	return false
}
