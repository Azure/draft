package safeguards

import (
	"fmt"
	"os"

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

	// Validate the chart
	if isHelm, err := chartutil.IsChartDir(chartDir); err != nil {
		return isHelm, fmt.Errorf("chart validation failed: %s", err)
	}

	return true, nil
}
