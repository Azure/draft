package web

import (
	"bytes"
	"os"

	"github.com/Azure/draft/pkg/filematches"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	parentDir               = "."
	deployNameToServiceYaml = map[string]*service{
		"helm":      {file: "charts/values.yaml", annotation: "service.annotations"},
		"kustomize": {file: "base/service.yaml", annotation: "metadata.annotations"},
	}
	// for testing purposes
	// deployType = "kustomize"
)

type service struct {
	file       string
	annotation string
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

	log.Debug("Loading config...")
	servicePath := deployNameToServiceYaml[deployType].file

	serviceBytes, err := os.ReadFile(servicePath)
	if err != nil {
		return err
	}

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(serviceBytes)); err != nil {
		return err
	}

	viper.Set(deployNameToServiceYaml[deployType].annotation, annotations)

	log.Debug("Writing new configuration to manifest...")
	if err := viper.WriteConfigAs(servicePath); err != nil {
		return err
	}

	return nil
}
