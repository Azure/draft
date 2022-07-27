package config

import (
	"errors"
	"fmt"
	"github.com/Azure/draft/pkg/consts"
	"github.com/Azure/draft/pkg/filematches"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart/loader"
	"path"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
)

type AddonConfig struct {
	DraftConfig `yaml:",inline"`
	References  map[string][]reference `yaml:"references"`

	deployType string
}

type reference struct {
	Name string
	Path string
}

type Reference interface {
	GetReferenceVariables([]reference) map[string]string
}

func (ac *AddonConfig) getDeployType(dest string) (string, error) {
	if ac.deployType != "" {
		return ac.deployType, nil
	}
	return filematches.FindDraftDeploymentFiles(dest)
}

func (ac *AddonConfig) GetAddonDestPath(dest string) (string, error) {
	deployType, err := ac.getDeployType(dest)
	if err != nil {
		return "", err
	}
	return path.Join(dest, consts.DeploymentFilePaths[deployType]), err
}

func (ac *AddonConfig) GetReferenceMap(dest string) (map[string]string, error) {
	referenceMap := make(map[string]string)

	deployType, err := ac.getDeployType(dest)

	for referenceName, references := range ac.References {
		switch deployType {
		case "helm":
			if err = getHelmReferenceMap(referenceName, dest, references, referenceMap); err != nil {
				return nil, err
			}

		case "kustomize":
			if err = getKustomizeReferenceMap(referenceName, dest, references, referenceMap); err != nil {
				return nil, err
			}

		case "manifests":
			if err = getManifestReferenceMap(referenceName, dest, references, referenceMap); err != nil {
				return nil, err
			}
		}
	}

	return referenceMap, err
}

// TODO: should consolidate all deployTypes into single interface to abstract the implementations
func getHelmReferenceMap(referenceName, dest string, references []reference, referenceMap map[string]string) error {
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

func getKustomizeReferenceMap(referenceName, dest string, references []reference, referenceMap map[string]string) error {
	kustomizer := krusty.MakeKustomizer(&krusty.Options{PluginConfig: &types.PluginConfig{}})
	production, err := kustomizer.Run(filesys.FileSystemOrOnDisk{}, dest+"/overlays/production/")
	if err != nil {
		return err
	}
	rNodes := production.ToRNodeSlice()

	return getNativeRefMap(rNodes, referenceName, references, referenceMap)
}

func getManifestReferenceMap(referenceName, dest string, references []reference, referenceMap map[string]string) error {
	serviceYaml, err := yaml.ReadFile(dest + "/manifests/service.yaml")
	if err != nil {
		return err
	}
	rNodes := make([]*yaml.RNode, 0)
	rNodes = append(rNodes, serviceYaml)
	return getNativeRefMap(rNodes, referenceName, references, referenceMap)
}

func getNativeRefMap(referenceNodes []*yaml.RNode, referenceName string, references []reference, referenceMap map[string]string) error {
	for _, reference := range references {
		refStr := getRef(referenceNodes, consts.RefPathLookups[referenceName][reference.Path])
		if refStr == "" && strings.Contains(reference.Name, "namespace") {
			//hack for default namespace
			refStr = "default"
		} else if refStr == "" {
			return errors.New(fmt.Sprintf("reference %s not found", reference.Name))
		}

		referenceMap[reference.Name] = refStr
	}
	return nil
}

func getRef(rNodes []*yaml.RNode, lookupPath []string) string {
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
