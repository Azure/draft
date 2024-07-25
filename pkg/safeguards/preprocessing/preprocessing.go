package preprocessing

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
			return RenderHelmChart(false, manifestsPath, tempDir)
		} else if isKustomize(true, manifestsPath) {
			return RenderKustomizeManifest(manifestsPath, tempDir)
		} else {
			manifestFiles, err = GetManifestFilesFromDir(manifestsPath)
			return manifestFiles, err
		}
	} else if IsYAML(manifestsPath) { // path points to a file
		if isHelm(false, manifestsPath) {
			return RenderHelmChart(true, manifestsPath, tempDir)
		} else if isKustomize(false, manifestsPath) {
			return RenderKustomizeManifest(manifestsPath, tempDir)
		} else {
			manifestFiles = append(manifestFiles, sgTypes.ManifestFile{
				Name: filepath.Base(manifestsPath),
				Path: manifestsPath,
			})
		}
		return manifestFiles, nil
	} else {
		return nil, fmt.Errorf("expected at least one .yaml or .yml file within given path")
	}
}

// getManifestFiles uses filepath.Walk to retrieve a list of the manifest files within a directory of .yaml files
func GetManifestFilesFromDir(p string) ([]sgTypes.ManifestFile, error) {
	var manifestFiles []sgTypes.ManifestFile

	err := filepath.Walk(p, func(walkPath string, info fs.FileInfo, err error) error {
		manifest := sgTypes.ManifestFile{}
		// skip when walkPath is just given path and also a directory
		if p == walkPath && info.IsDir() {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking path %s with error: %w", walkPath, err)
		}

		if !info.IsDir() && info.Name() != "" && IsYAML(walkPath) {
			log.Debugf("%s is not a directory, appending to manifestFiles", info.Name())

			manifest.Name = info.Name()
			manifest.Path = walkPath
			manifestFiles = append(manifestFiles, manifest)
		} else if !IsYAML(p) {
			log.Debugf("%s is not a manifest file, skipping...", info.Name())
		} else {
			log.Debugf("%s is a directory, skipping...", info.Name())
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not walk directory: %w", err)
	}
	if len(manifestFiles) == 0 {
		return nil, fmt.Errorf("no manifest files found within given path")
	}

	return manifestFiles, nil
}

// Given a Helm chart directory or file, renders all templates and writes them to the specified directory
func RenderHelmChart(isFile bool, mainChartPath, tempDir string) ([]sgTypes.ManifestFile, error) {
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
			outputFilePath := filepath.Join(tempDir, filepath.Base(renderedPath))
			if err := os.WriteFile(outputFilePath, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write manifest file: %s", err)
			}
			manifestFiles = append(manifestFiles, sgTypes.ManifestFile{Name: filepath.Base(renderedPath), Path: outputFilePath})
		}
	}

	return manifestFiles, nil
}

// CreateTempDir creates a temporary directory on the user's file system for rendering templates
func CreateTempDir(p string) error {
	err := os.MkdirAll(p, 0755)
	if err != nil {
		log.Fatal(err)
	}

	return err
}

// Given a kustomization manifest file within kustomizationPath, RenderKustomizeManifest will render templates out to tempDir
func RenderKustomizeManifest(kustomizationPath, tempDir string) ([]sgTypes.ManifestFile, error) {
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
	kindMap := make(map[string]int)
	for _, res := range resMap.Resources() {
		yamlRes, err := res.AsYAML()
		if err != nil {
			return nil, fmt.Errorf("error converting resource to YAML: %s", err.Error())
		}

		// index of every kind of manifest for outputRenderPath
		kindMap[res.GetKind()] += 1
		outputRenderPath := filepath.Join(tempDir, strings.ToLower(res.GetKind())) + fmt.Sprintf("-%d.yaml", kindMap[res.GetKind()])

		err = kustomizeFS.WriteFile(outputRenderPath, yamlRes)
		if err != nil {
			return nil, fmt.Errorf("error writing yaml resource: %s", err.Error())
		}

		// write yamlRes to dir
		manifestFiles = append(manifestFiles, sgTypes.ManifestFile{
			Name: res.GetName(),
			Path: outputRenderPath,
		})
	}

	return manifestFiles, nil
}
