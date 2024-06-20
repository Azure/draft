package safeguards

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chartutil"
)

// DirectoryType represents the type of directory.
type DirectoryType int

const (
	Unknown   DirectoryType = iota // Default value if type cannot be determined
	Helm                           // Represents a Helm chart directory
	Kustomize                      // Represents a Kustomize configuration directory
)

func CreateTempDir() (string, error) {
	log.Infof("Creating temporary directory for rendered kustomize/helm manifests")
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Errorf("Error creating temporary directory to store rendered kustomize/helm manifests: %s", err)
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	return tempDir, nil
}

func GetDirectoryType(dirPath string) DirectoryType {
	if isHelm, _ := isHelmChart(dirPath); isHelm {
		return Helm
	}
	// } else if isKustomizeConfig(dirPath) {
	// 		return Kustomize
	// }

	return Unknown
}

func isHelmChart(chartDir string) (bool, error) {
	// Check if the directory exists
	if _, err := os.Stat(chartDir); os.IsNotExist(err) {
		return false, fmt.Errorf("directory does not exist: %s", chartDir)
	}

	// // Load the chart
	// _, err := loader.Load(chartDir)
	// if err != nil {
	// 	return fmt.Errorf("failed to load chart: %s", err)
	// }

	// Validate the chart
	if isHelm, err := chartutil.IsChartDir(chartDir); err != nil {
		return isHelm, fmt.Errorf("chart validation failed: %s", err)
	}

	return true, nil
}

// isHelmChart checks if the given directory is a valid Helm chart with required components.
func isHelmChart2(chartDir string) error {
	chartFile := filepath.Join(chartDir, "Chart.yaml")
	valuesFile := filepath.Join(chartDir, "Values.yaml")
	templatesDir := filepath.Join(chartDir, "templates")

	// Check Chart.yaml
	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		return fmt.Errorf("missing Chart.yaml: %w", err)
	}

	// Check values.yaml
	if _, err := os.Stat(valuesFile); os.IsNotExist(err) {
		return fmt.Errorf("missing values.yaml: %w", err)
	}

	// Check templates directory
	fileInfo, err := os.Stat(templatesDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("missing templates directory: %w", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("templates is not a directory")
	}

	// If we are only checking for one helm chart, we can just make sure the number of components in the dir == 3

	// Check for any other files or directories in the chart directory
	entries, err := os.ReadDir(chartDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	if len(entries) != 3 {
		return fmt.Errorf("other entries are present in your directory apart from your helm chart, please remove all other files")
	}
	// for _, entry := range entries {
	// 	if entry.Name() != "Chart.yaml" && entry.Name() != "values.yaml" && entry.Name() != "templates" {
	// 		return false, fmt.Errorf("unexpected file or directory in Helm chart: %s", entry.Name())
	// 	}
	// }

	return nil
}

// func IsKustomizeConfig(input string) bool {

// }
