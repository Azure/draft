package web

import (
	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/types"
	log "github.com/sirupsen/logrus"
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
	log.Debugf("editing service yaml for deployType: %s", deployType)
	switch deployType {
	case "helm":
		return updateDeploymentAnnotations(&types.HelmProductionYaml{}, filePath, annotations)
	}

	return updateDeploymentAnnotations(&types.ServiceYaml{}, filePath, annotations)
}

func updateDeploymentAnnotations[K types.ServiceManifest](deploy K, filePath string, annotations map[string]string) error {
	if err := deploy.LoadFromFile(filePath); err != nil {
		return err
	}

	deploy.SetAnnotations(annotations)
	deploy.SetServiceType("ClusterIP")

	return deploy.WriteToFile(filePath)
}
