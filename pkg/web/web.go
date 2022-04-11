package web

import (
	"bytes"
	"os"

	//"github.com/Azure/draftv2/pkg/filematches"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)


func updateServiceFile() error {
	// 	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	// 	if err != nil {
	// 		return err
	// 	}

	log.Debug("Loading config...")
	// TODO: figure out actual path to service.yaml
	serviceBytes, err := os.ReadFile("deployTypes/helm")
	if err != nil {
		return err
	}

	viper.SetConfigFile("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(serviceBytes)); err != nil {
		return err
	}

	var serviceConfig map[string]string
	if err = viper.Unmarshal(&serviceConfig); err != nil {
		return err
	}

	// logic to change service yaml as needed for ingress

	return nil
}