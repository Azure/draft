package web

import (
	"bytes"
	"encoding/json"
	"fmt"
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
)

type Metadata struct {
	Name string `json:"name"`
	Annotations []string `json:"annotations"`
}

func UpdateServiceFile() error {
	// 	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	// 	if err != nil {
	// 		return err
	// 	}
	deployType := "kustomize"

	// TODO: change annotations in values.yaml for helm to []?
	// or change to string then find and replace annotations: {}

	log.Debug("Loading config...")
	servicePath := parentDir + "/" + deployNameToServiceYaml[deployType] + ".yaml"
	serviceBytes, err := os.ReadFile(servicePath)
	if err != nil {
		return err
	}

	//fmt.Printf(string(serviceBytes))

	viper.SetConfigFile("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(serviceBytes)); err != nil {
		return err
	}

	var serviceConfig map[string]interface{}
	if err = viper.Unmarshal(&serviceConfig); err != nil {
		return err
	}

	
	// logic to change service yaml as needed for ingress
	// metadata := new(Metadata)
	// metadata.Name = "my-app"
	// metadata.Annotations = []string{"kubernetes.azure.com/ingress-host", "kubernetes.azure.com/tls-cert-keyvault-uri"}

	serviceConfig["metadata"] = `{name: "my-app", annotations: ["kubernetes.azure.com/ingress-host", "kubernetes.azure.com/tls-cert-keyvault-uri"]}`

	data, err := json.Marshal(&serviceConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(servicePath, data, 0644); err != nil {
		return err
	}

	fmt.Print(serviceConfig["metadata"])

	return nil
}