package preprocessing

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Returns values from values.yaml and release options specified in values.yaml
func getValues(chart *chart.Chart, valuesPath string, opt chartutil.ReleaseOptions, dirName string) (chartutil.Values, error) {
	// Load values file
	valuesFile, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %s", err)
	}

	vals := map[string]interface{}{}
	if err := yaml.Unmarshal(valuesFile, &vals); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml: %s", err)
	}

	mergedValues, err := getReleaseOptions(chart, vals, opt, dirName)
	return mergedValues, err
}

func getReleaseOptions(chart *chart.Chart, vals map[string]interface{}, opt chartutil.ReleaseOptions, dirName string) (chartutil.Values, error) {
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
			rName, ok := vals["releaseName"].(string)
			if !ok || rName == "" {
				releaseName = dirName
			} else {
				releaseName = rName
			}
		}
		if opt.Namespace != "" {
			releaseNamespace = opt.Namespace
		} else {
			rNamespace, ok := vals["releaseNamespace"].(string)
			if !ok || rNamespace == "" {
				releaseNamespace = dirName
			} else {
				releaseNamespace = rNamespace
			}
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
