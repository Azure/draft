package cmdhelpers

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/consts"
	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/prompts"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart/loader"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type referenceResource struct {
	Name string
	Path string
}

var ReferenceResources map[string][]referenceResource = map[string][]referenceResource{
	"service": {
		{Name: "namespace", Path: "metadata.namespace"},
		{Name: "name", Path: "metadata.name"},
		{Name: "service-port", Path: "spec.ports.port"},
	},
}

func PromptAddonValues(dest string, addonConfig *config.DraftConfig) error {
	err := prompts.RunPromptsFromConfigWithSkips(addonConfig)
	if err != nil {
		return err
	}
	log.Debug("got user inputs")

	deployType, err := getDeployType(dest)
	if err != nil {
		return err
	}

	referenceMap, err := GetReferenceValueMap(dest, deployType)
	if err != nil {
		return err
	}
	log.Debug("got reference map")
	// merge maps
	for refName, refVal := range referenceMap {
		// check for key collision
		if _, err := addonConfig.GetVariable(refName); err == nil {
			return errors.New("variable name collision between references and DraftConfig")
		}
		if strings.Contains(strings.ToLower(refName), "namespace") && refVal == "" {
			refVal = "default" //hack here to have explicit namespacing, probably a better way to do this
		}
		addonConfig.SetVariable(refName, refVal)
	}

	return nil
}

func GetAddonDestPath(dest string) (string, error) {
	deployType, err := getDeployType(dest)
	if err != nil {
		return "", err
	}
	return path.Join(dest, consts.DeploymentFilePaths[deployType]), err
}

func getDeployType(dest string) (string, error) {
	deploymentType, err := filematches.FindDraftDeploymentFiles(dest)
	log.Debugf("found deployment type: %s", deploymentType)
	return deploymentType, err
}

func GetReferenceValueMap(deployType, dest string) (map[string]string, error) {
	referenceMap := make(map[string]string)
	var err error
	for referenceName, referenceResources := range ReferenceResources {
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
				consts.HelmReferencePathMapping[referenceName][reference.Path], ".Config.GetVariableValue \"APPNAME\"", chart.Name())
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
			return fmt.Errorf("referenceResource %s not found", reference.Name)
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

		refYNode := ref.YNode() // Convert to go yaml.node to extract values without trailing newlines
		refStr := refYNode.Value
		log.Debugf("found ref: %s", refStr)
		return refStr
	}
	return ""
}
