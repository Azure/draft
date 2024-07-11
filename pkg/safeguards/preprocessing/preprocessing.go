package preprocessing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/safeguards"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/engine"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Given a Helm chart directory or file, renders all templates and writes them to the specified directory
func RenderHelmChart(isFile bool, mainChartPath, tempDir string) ([]safeguards.ManifestFile, error) {
	if isFile { // Get the directory that the Chart.yaml lives in
		mainChartPath = filepath.Dir(mainChartPath)
	}

	mainChart, err := loader.Load(mainChartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load main chart: %s", err)
	}

	loadedCharts := make(map[string]*chart.Chart) // map of chart path to chart object
	loadedCharts[mainChartPath] = mainChart

	// Load subcharts and dependencies
	for _, dep := range mainChart.Metadata.Dependencies {
		// Resolve the chart path based on the main chart's directory
		chartPath := filepath.Join(mainChartPath, dep.Repository[len("file://"):])
		chartPath = filepath.Clean(chartPath)

		subChart, err := loader.Load(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load chart: %s", err)
		}
		loadedCharts[chartPath] = subChart
	}

	var manifestFiles []safeguards.ManifestFile
	for chartPath, chart := range loadedCharts {
		valuesPath := filepath.Join(chartPath, "values.yaml") // Enforce that values.yaml must be at same level as Chart.yaml
		mergedValues, err := getValues(chart, valuesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load values: %s", err)
		}
		e := engine.Engine{Strict: true}
		renderedFiles, err := e.Render(chart, mergedValues)
		if err != nil {
			return nil, fmt.Errorf("failed to render chart: %s", err)
		}

		// Write each rendered file to the output directory with the same name as in templates/
		for renderedPath, content := range renderedFiles {
			outputFilePath := filepath.Join(tempDir, filepath.Base(renderedPath))
			if err := os.WriteFile(outputFilePath, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write manifest file: %s", err)
			}
			manifestFiles = append(manifestFiles, safeguards.ManifestFile{Name: filepath.Base(renderedPath), Path: outputFilePath})
		}
	}

	return manifestFiles, nil
}

// CreateTempDir creates a temporary directory on the user's file system for rendering templates
func CreateTempDir(p string) string {
	dir, err := os.MkdirTemp(p, "prefix")
	if err != nil {
		log.Fatal(err)
	}

	return dir
}

// IsKustomize checks whether a given path should be treated as a kustomize project
func IsKustomize(p string) bool {
	var err error
	if safeguards.IsYAML(p) {
		return strings.Contains(p, "kustomization.yaml")
	} else if _, err = os.Stat(filepath.Join(p, "kustomization.yaml")); err == nil {
		return true
	} else if _, err = os.Stat(filepath.Join(p, "kustomization.yml")); err == nil {
		return true
	}
	return false
}

// Given a kustomization manifest file within kustomizationPath, RenderKustomizeManifest will render templates out to tempDir
func RenderKustomizeManifest(kustomizationPath, tempDir string) ([]safeguards.ManifestFile, error) {
	log.Debugf("Rendering kustomization.yaml...")

	options := &krusty.Options{
		Reorder:           "none",
		AddManagedbyLabel: true,
		LoadRestrictions:  types.LoadRestrictionsRootOnly,
		PluginConfig:      &types.PluginConfig{},
	}
	k := krusty.MakeKustomizer(options)

	// Run the build to generate the manifests
	kustomizeFS := filesys.MakeFsOnDisk()
	resMap, err := k.Run(kustomizeFS, kustomizationPath)
	if err != nil {
		return nil, fmt.Errorf("Error building manifests: %s\n", err.Error())
	}

	// Output the manifests
	var manifestFiles []safeguards.ManifestFile
	kindMap := make(map[string]int)
	for _, res := range resMap.Resources() {
		yamlRes, err := res.AsYAML()
		if err != nil {
			return nil, fmt.Errorf("Error converting resource to YAML: %s\n", err.Error())
		}

		// index of every kind of manifest for outputRenderPath
		kindMap[res.GetKind()] += 1
		outputRenderPath := filepath.Join(tempDir, strings.ToLower(res.GetKind())) + fmt.Sprintf("-%d.yaml", kindMap[res.GetKind()])

		err = kustomizeFS.WriteFile(outputRenderPath, yamlRes)
		if err != nil {
			return nil, fmt.Errorf("Error writing yaml resource: %s\n", err.Error())
		}

		// write yamlRes to dir
		manifestFiles = append(manifestFiles, safeguards.ManifestFile{
			Name: res.GetName(),
			Path: outputRenderPath,
		})
	}

	return manifestFiles, nil
}
