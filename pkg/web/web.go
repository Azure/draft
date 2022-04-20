package web

import (
	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/workflows"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	parentDir               = "."
	deployNameToServiceYaml = map[string]*service{
		"helm":      {file: "charts/production.yaml"},
		"kustomize": {file: "overlays/production/service.yaml"},
		"manifests": {file: "manifests/service.yaml"},
	}
)

type service struct {
	file string
}

type ServiceAnnotations struct {
	Host string
	Cert string
}

func UpdateServiceFile(sa *ServiceAnnotations, dest string) error {
	annotations := map[string]string{
		"kubernetes.azure.com/ingress-host":          sa.Host,
		"kubernetes.azure.com/tls-cert-keyvault-uri": sa.Cert,
	}

	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return err
	}

	servicePath := dest + "/" + deployNameToServiceYaml[deployType].file
	log.Debug("Writing new configuration to manifest...")

	return updateServiceAnnotationsForDeployment(servicePath, deployType, annotations)
}

func updateServiceAnnotationsForDeployment(filePath, deployType string, annotations map[string]string) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var editedYaml []byte

	log.Debugf("editing service yaml for deployType: %s", deployType)
	switch deployType {
	case "helm":
		var deploy workflows.HelmProductionYaml
		editedYaml, err = updateDeploymentAnnotations(&deploy, file, annotations)
		if err != nil {
			return err
		}
	default:
		var deploy workflows.ServiceYaml
		editedYaml, err = updateDeploymentAnnotations(&deploy, file, annotations)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filePath, editedYaml, 0644)
}

func updateDeploymentAnnotations[K workflows.ServiceManifest](deploy K, file []byte, annotations map[string]string) ([]byte, error) {
	err := yaml.Unmarshal(file, deploy)
	if err != nil {
		return nil, err
	}

	deploy.SetAnnotations(annotations)
	deploy.SetServiceType("ClusterIP")

	return yaml.Marshal(deploy)
}
