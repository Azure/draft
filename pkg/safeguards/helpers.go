package safeguards

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"

	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/rego"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/target"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Azure/draft/pkg/safeguards/preprocessing"
	"github.com/Azure/draft/pkg/safeguards/types"
)

// Given a path, will determine if it's Kustomize, Helm, a directory of manifests, or a single manifest
func GetManifestFiles(manifestsPath string, opt chartutil.ReleaseOptions) ([]types.ManifestFile, error) {
	isDir, err := IsDirectory(manifestsPath)
	if err != nil {
		return nil, fmt.Errorf("not a valid file or directory: %w", err)
	}

	var manifestFiles []types.ManifestFile
	if isDir {
		// check if Helm or Kustomize dir
		if isHelm(true, manifestsPath) {
			return preprocessing.RenderHelmChart(false, manifestsPath, opt)
		} else if isKustomize(true, manifestsPath) {
			return preprocessing.RenderKustomizeManifest(manifestsPath)
		} else {
			manifestFiles, err = GetManifestFilesFromDir(manifestsPath)
			return manifestFiles, err
		}
	} else if IsYAML(manifestsPath) { // path points to a file
		if isHelm(false, manifestsPath) {
			return preprocessing.RenderHelmChart(true, manifestsPath, opt)
		} else if isKustomize(false, manifestsPath) {
			return preprocessing.RenderKustomizeManifest(manifestsPath)
		} else {
			byteContent, err := os.ReadFile(manifestsPath)
			if err != nil {
				return nil, fmt.Errorf("could not read file %s: %s", manifestsPath, err)
			}
			manifestFiles = append(manifestFiles, types.ManifestFile{
				Name:            path.Base(manifestsPath),
				ManifestContent: byteContent,
			})
		}
		return manifestFiles, nil
	} else {
		return nil, fmt.Errorf("expected at least one .yaml or .yml file within given path")
	}
}

// GetManifestFilesFromDir uses filepath.Walk to retrieve a list of the manifest files within a directory of .yaml files
func GetManifestFilesFromDir(p string) ([]types.ManifestFile, error) {
	var manifestFiles []types.ManifestFile

	err := filepath.Walk(p, func(walkPath string, info fs.FileInfo, err error) error {
		manifest := types.ManifestFile{}
		// skip when walkPath is just given path and also a directory
		if p == walkPath && info.IsDir() {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error walking path %s with error: %w", walkPath, err)
		}

		if !info.IsDir() && info.Name() != "" && IsYAML(walkPath) {
			log.Debugf("%s is not a directory, appending to manifestFiles", info.Name())

			byteContent, err := os.ReadFile(walkPath)
			if err != nil {
				return fmt.Errorf("could not read file %s: %s", walkPath, err)
			}
			manifest.Name = info.Name()
			manifest.ManifestContent = byteContent
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

// retrieves the constraint client that does all rego code related operations
func getConstraintClient() (*constraintclient.Client, error) {
	driver, err := rego.New()
	if err != nil {
		return nil, fmt.Errorf("could not create rego driver: %w", err)
	}

	c, err := constraintclient.NewClient(constraintclient.Targets(&target.K8sValidationTarget{}), constraintclient.Driver(driver))
	if err != nil {
		return nil, fmt.Errorf("could not create constraint client: %w", err)
	}

	return c, nil
}

// sorts the list of supported safeguards versions and returns the last item in the list
func getLatestSafeguardsVersion() string {
	semver.Sort(types.SupportedVersions)
	return types.SupportedVersions[len(types.SupportedVersions)-1]
}

func updateSafeguardPaths(safeguardList *[]types.Safeguard) {
	for _, sg := range *safeguardList {
		sg.TemplatePath = fmt.Sprintf("%s/%s/%s", types.SelectedVersion, sg.Name, types.TemplateFileName)
		sg.ConstraintPath = fmt.Sprintf("%s/%s/%s", types.SelectedVersion, sg.Name, types.ConstraintFileName)
	}
}

// adds Safeguard_CRIP to full list of Safeguards
func AddSafeguardCRIP() {
	fc.Safeguards = append(fc.Safeguards, types.Safeguard_CRIP)
}

// loads constraint templates, constraints into constraint client
func loadConstraintTemplates(ctx context.Context, c *constraintclient.Client, constraintTemplates []*templates.ConstraintTemplate) error {
	// AddTemplate adds the template source code to OPA and registers the CRD with the client for
	// schema validation on calls to AddConstraint. On error, the responses return value
	// will still be populated so that partial results can be analyzed.
	for _, ct := range constraintTemplates {
		_, err := c.AddTemplate(ctx, ct)
		if err != nil {
			return fmt.Errorf("could not add template: %w", err)
		}
	}

	return nil
}

func loadConstraints(ctx context.Context, c *constraintclient.Client, constraints []*unstructured.Unstructured) error {
	// AddConstraint validates the constraint and, if valid, inserts it into OPA.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.
	for _, con := range constraints {
		_, err := c.AddConstraint(ctx, con)
		if err != nil {
			return fmt.Errorf("could not add constraint: %w", err)
		}
	}

	return nil
}

func loadManifestObjects(ctx context.Context, c *constraintclient.Client, objects []*unstructured.Unstructured) error {
	// AddData inserts the provided data into OPA for every target that can handle the data.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.
	for _, o := range objects {
		_, err := c.AddData(ctx, o)
		if err != nil {
			return fmt.Errorf("could not add data: %w", err)
		}
	}

	return nil
}

// IsDirectory determines if a file represented by path is a directory or not
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

// IsYAML determines if a file is of the YAML extension or not
func IsYAML(path string) bool {
	return filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"
}

// getObjectViolations executes validation on manifests based on loaded constraint templates and returns a map of manifest name to list of objectViolations
func getObjectViolations(ctx context.Context, c *constraintclient.Client, objects []*unstructured.Unstructured) (map[string][]string, error) {
	// Review makes sure the provided object satisfies all stored constraints.
	// On error, the responses return value will still be populated so that
	// partial results can be analyzed.

	var results = make(map[string][]string) // map of object name to slice of objectViolations

	for _, o := range objects {
		objectViolations := []string{}
		log.Debugf("Reviewing %s...", o.GetName())
		res, err := c.Review(ctx, o)
		if err != nil {
			return results, fmt.Errorf("could not review objects: %w", err)
		}

		for _, v := range res.ByTarget {
			for _, result := range v.Results {
				if result.Msg != "" {
					objectViolations = append(objectViolations, result.Msg)
				}
			}
		}

		if len(objectViolations) > 0 {
			results[o.GetName()] = objectViolations
		}
	}

	return results, nil
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
