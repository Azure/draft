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
	deployNameToServiceYaml = map[string]string{
		"helm": "charts/values",
		"kustomize": "base/service",

	}
	annotations = map[string]string{
		"kubernetes.azure.com/ingress-host": "placeholder",
		"kubernetes.azure.com/tls-cert-keyvault-uri": "placeholder",
	}
)

func UpdateServiceFile() error {
	// 	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	// 	if err != nil {
	// 		return err
	// 	}

	// for testing purposes
	deployType := "kustomize"

	// TODO: change annotations in values.yaml for helm

	log.Debug("Loading config...")
	servicePath := parentDir + "/" + deployNameToServiceYaml[deployType] + ".yaml"
	serviceBytes, err := os.ReadFile(servicePath)
	if err != nil {
		return err
	}

	if err := viper.ReadConfig(bytes.NewBuffer(serviceBytes)); err != nil {
		return err
	}

	viper.Set("metadata.annotations", annotations)

	if err := viper.WriteConfigAs("./base/test.yaml"); err != nil {
		return err
	}

	return nil
}