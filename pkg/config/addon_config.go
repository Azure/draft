package config

import (
	"github.com/Azure/draft/pkg/consts"
	"github.com/Azure/draft/pkg/filematches"
	"helm.sh/helm/v3/pkg/chart/loader"
	"strings"
)

type AddonConfig struct {
	DraftConfig
	References map[string][]reference
}

type reference struct {
	Name string
	Path string
}

type Reference interface {
	GetReferenceVariables([]reference) map[string]string
}

func (ac *AddonConfig) GetReferenceMap(dest string) (map[string]string, error) {
	referenceMap := make(map[string]string)

	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return nil, err
	}
	for referenceName, references := range ac.References {
		switch deployType {
		case "helm":
			if err = getHelmReferenceMap(referenceName, dest, references, referenceMap); err != nil {
				return nil, err
			}

		case "kustomize":
			return referenceMap, nil
		}
	}

	return referenceMap, err
}

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
