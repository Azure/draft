package preprocessing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Returns values from values.yaml and release options specified in values.yaml
func getValues(chart *chart.Chart, valuesPath string, opt chartutil.ReleaseOptions, containingDir string) (chartutil.Values, error) {
	// Load values file
	valuesFile, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %s", err)
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesFile, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %s", err)
	}

	mergedValues, err := getReleaseOptions(chart, vals, opt, containingDir)
	return mergedValues, err
}

func getReleaseOptions(chart *chart.Chart, vals map[string]interface{}, opt chartutil.ReleaseOptions, containingDir string) (chartutil.Values, error) {
	// Extract release options from values

	var options chartutil.ReleaseOptions
	if opt.Name != "" && opt.Namespace != "" {
		options = opt
	} else {
		var releaseName string
		var releaseNamespace string
		if opt.Name != "" {
			releaseName = opt.Name
		} else {
			releaseName = containingDir
		}
		if opt.Namespace != "" {
			releaseNamespace = opt.Namespace
		} else {
			releaseNamespace = containingDir
		}

		options = chartutil.ReleaseOptions{
			Name:      releaseName,
			Namespace: releaseNamespace,
		}
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
	var chartPaths []string // Used to define what a valid helm chart looks like. Currently, presence of Chart.yaml/.yml.

	if isDir {
		chartPaths = []string{filepath.Join(path, "Chart.yaml")}
		chartPaths = append(chartPaths, filepath.Join(path, "Chart.yml"))
	} else {
		if filepath.Base(path) != "Chart.yaml" && filepath.Base(path) != "Chart.yml" {
			return false
		}
		chartPaths = []string{path}
	}

	for _, path := range chartPaths {
		_, err := os.Stat(path)
		if err == nil { //Found the file, it's a valid helm chart
			return true
		}
	}

	return false
}

// IsYAML determines if a file is of the YAML extension or not
func IsYAML(path string) bool {
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}

// IsDirectory determines if a file represented by path is a directory or not
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
