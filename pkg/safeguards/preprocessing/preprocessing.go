package preprocessing

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Azure/draft/pkg/safeguards"
	sgTypes "github.com/Azure/draft/pkg/safeguards/types"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/engine"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

// Given a path, will determine if it's Kustomize, Helm, a directory of manifests, or a single manifest
func GetManifestFiles(manifestsPath string) ([]sgTypes.ManifestFile, error) {
	isDir, err := IsDirectory(manifestsPath)
	if err != nil {
		return nil, fmt.Errorf("not a valid file or directory: %w", err)
	}

	var manifestFiles []sgTypes.ManifestFile
	if isDir {
		// check if Helm or Kustomize dir
		if isHelm(true, manifestsPath) {
			return RenderHelmChart(false, manifestsPath)
		} else if isKustomize(true, manifestsPath) {
			return RenderKustomizeManifest(manifestsPath)
		} else {
			manifestFiles, err = safeguards.GetManifestFilesFromDir(manifestsPath)
			return manifestFiles, err
		}
	} else if IsYAML(manifestsPath) { // path points to a file
		if isHelm(false, manifestsPath) {
			return RenderHelmChart(true, manifestsPath)
		} else if isKustomize(false, manifestsPath) {
			return RenderKustomizeManifest(manifestsPath)
		} else {
			byteContent, err := os.ReadFile(manifestsPath)
			if err != nil {
				return nil, fmt.Errorf("could not read file %s: %s", manifestsPath, err)
			}
			manifestFiles = append(manifestFiles, sgTypes.ManifestFile{
				Name:            path.Base(manifestsPath),
				ManifestContent: byteContent,
			})
		}
		return manifestFiles, nil
	} else {
		return nil, fmt.Errorf("expected at least one .yaml or .yml file within given path")
	}
}

// Given a Helm chart directory or file, renders all templates and writes them to the specified directory
func RenderHelmChart(isFile bool, mainChartPath string) ([]sgTypes.ManifestFile, error) {
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

	var manifestFiles []sgTypes.ManifestFile
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
			byteContent := []byte(content)
			manifestFiles = append(manifestFiles, sgTypes.ManifestFile{Name: filepath.Base(renderedPath), ManifestContent: byteContent})
		}
	}

	return manifestFiles, nil
}

// Given a kustomization manifest file within kustomizationPath, RenderKustomizeManifest will return render templates
func RenderKustomizeManifest(kustomizationPath string) ([]sgTypes.ManifestFile, error) {
	log.Debugf("Rendering kustomization.yaml...")
	if IsYAML(kustomizationPath) {
		kustomizationPath = filepath.Dir(kustomizationPath)
	}

	options := &krusty.Options{
		Reorder:          "none",
		LoadRestrictions: types.LoadRestrictionsRootOnly,
		PluginConfig:     &types.PluginConfig{},
	}
	k := krusty.MakeKustomizer(options)

	// Run the build to generate the manifests
	kustomizeFS := filesys.MakeFsOnDisk()
	resMap, err := k.Run(kustomizeFS, kustomizationPath)
	if err != nil {
		return nil, fmt.Errorf("error building manifests: %s", err.Error())
	}

	// Output the manifests
	var manifestFiles []sgTypes.ManifestFile
	for _, res := range resMap.Resources() {
		yamlRes, err := res.AsYAML()
		if err != nil {
			return nil, fmt.Errorf("error converting resource to YAML: %s", err.Error())
		}

		// write yamlRes to dir
		manifestFiles = append(manifestFiles, sgTypes.ManifestFile{
			Name:            res.GetName(),
			ManifestContent: yamlRes,
		})
	}

	return manifestFiles, nil
}
