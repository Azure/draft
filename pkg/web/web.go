package web

import (
	"bytes"
	"os"

	//"github.com/Azure/draftv2/pkg/filematches"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	parentDir = "."
	deployNameToServiceYaml = map[string]*service{
		"helm": {file: "charts/values", annotation: "service.annotations"},
		"kustomize": {file: "base/service", annotation: "metadata.annotations"},

	}
	annotations = map[string]string{
		"kubernetes.azure.com/ingress-host": "placeholder",
		"kubernetes.azure.com/tls-cert-keyvault-uri": "placeholder",
	}
)

type service struct {
	file string
	annotation string
}

func UpdateServiceFile() error {
	// 	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	// 	if err != nil {
	// 		return err
	// 	}

	// for testing purposes
	deployType := "helm"

	log.Debug("Loading config...")
	servicePath := parentDir + "/" + deployNameToServiceYaml[deployType].file + ".yaml"
	serviceBytes, err := os.ReadFile(servicePath)
	if err != nil {
		return err
	}

	
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