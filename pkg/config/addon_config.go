package config

import (
	"errors"
	"fmt"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart/loader"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/Azure/draft/pkg/consts"
	"github.com/Azure/draft/pkg/filematches"
)

// AddonConfig is a struct that extends the base DraftConfig to allow for the Referencing previously generated
// k8s objects. This allows an addon creator to reference pre-entered data from the deployment files.
type AddonConfig struct {
	DraftConfig         `yaml:",inline"`
	ReferenceComponents map[string][]referenceResource `yaml:"references"`

	deployType string
}

type referenceResource struct {
	Name string
	Path string
}

type Reference interface {
	GetReferenceVariables([]referenceResource) map[string]string
}

func (ac *AddonConfig) getDeployType(dest string) (string, error) {
	if ac.deployType != "" {
		return ac.deployType, nil
	}
	deploymentType, err := filematches.FindDraftDeploymentFiles(dest)
	log.Debugf("found deployment type: %s", deploymentType)
	return deploymentType, err
}

func (ac *AddonConfig) GetAddonDestPath(dest string) (string, error) {
	deployType, err := ac.getDeployType(dest)
	if err != nil {
		return "", err
	}
	return path.Join(dest, consts.DeploymentFilePaths[deployType]), err
}

// GetReferenceValueMap extracts k8s object values into a mapping of template strings to k8s object value.
func (ac *AddonConfig) GetReferenceValueMap(dest string) (map[string]string, error) {
	referenceMap := make(map[string]string)

	deployType, err := ac.getDeployType(dest)

	for referenceName, referenceResources := range ac.ReferenceComponents {
		switch deployType {
		case "helm":
			if err = extractHelmValuesToMap(referenceName, dest, referenceResources, referenceMap); err != nil {
				return nil, err
			}

		case "kustomize":
			if err = extractKustomizeValuesToMap(referenceName, dest, referenceResources, referenceMap); err != nil {
				return nil, err
			}

		case "manifests":
			if err = extractManifestValuesToMap(referenceName, dest, referenceResources, referenceMap); err != nil {
				return nil, err
			}
		}
	}

	return referenceMap, err
}

// TODO: should consolidate all deployTypes into single interface to abstract the implementations
func extractHelmValuesToMap(referenceName, dest string, references []referenceResource, referenceMap map[string]string) error {
	chart, err := loader.Load(dest + "/charts/")
	if err != nil {
		return err
	}

	for _, reference := range references {
		referenceMap[reference.Name] =
			strings.ReplaceAll(
				consts.HelmReferencePathMapping[referenceName][reference.Path], "{{APPNAME}}", chart.Name())
	}

	return nil
}

func extractKustomizeValuesToMap(referenceName, dest string, references []referenceResource, referenceMap map[string]string) error {
	kustomizer := krusty.MakeKustomizer(&krusty.Options{PluginConfig: &types.PluginConfig{}})
	production, err := kustomizer.Run(filesys.FileSystemOrOnDisk{}, dest+"/overlays/production/")
	if err != nil {
		return err
	}
	rNodes := production.ToRNodeSlice()

	return extractNativeRefMap(rNodes, referenceName, references, referenceMap)
}

func extractManifestValuesToMap(referenceName, dest string, references []referenceResource, referenceMap map[string]string) error {
	serviceYaml, err := yaml.ReadFile(dest + "/manifests/service.yaml")
	if err != nil {
		return err
	}
	rNodes := make([]*yaml.RNode, 0)
	rNodes = append(rNodes, serviceYaml)
	return extractNativeRefMap(rNodes, referenceName, references, referenceMap)
}

func extractNativeRefMap(referenceNodes []*yaml.RNode, referenceName string, references []referenceResource, referenceMap map[string]string) error {
	for _, reference := range references {
		refStr := extractRef(referenceNodes, consts.RefPathLookups[referenceName][reference.Path])
		if refStr == "" && strings.Contains(reference.Name, "namespace") {
			//hack for default namespace
			refStr = "default"
		} else if refStr == "" {
			return errors.New(fmt.Sprintf("referenceResource %s not found", reference.Name))
		}

		referenceMap[reference.Name] = refStr
	}
	return nil
}

func extractRef(rNodes []*yaml.RNode, lookupPath []string) string {
	for _, rNode := range rNodes {
		ref, err := rNode.Pipe(yaml.Lookup(lookupPath...))
		if ref == nil || err != nil {
			continue
		}
		refStr, _ := ref.String()
		log.Debugf("found ref: %s", refStr)
		return refStr
	}
	return ""
}
