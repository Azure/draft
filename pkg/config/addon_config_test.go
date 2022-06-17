package config

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"testing"
)

func TestGetHelmReferenceMap(t *testing.T) {
	configBytes, err := ioutil.ReadFile("../addons/addons/azure/webapp_routing/draft_config.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	assert.Nil(t, err)

	refMap := make(map[string]string)
	err = getHelmReferenceMap("service", "../../test/templates/helm", addOnConfig.References["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetKustomizeReferenceMap(t *testing.T) {
	configBytes, err := ioutil.ReadFile("../addons/addons/azure/webapp_routing/draft_config.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	refMap := make(map[string]string)

	err = getKustomizeReferenceMap("service", "../../test/templates/kustomize", addOnConfig.References["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetManifestReferenceMap(t *testing.T) {
	configBytes, err := ioutil.ReadFile("../addons/addons/azure/webapp_routing/draft_config.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	refMap := make(map[string]string)

	err = getManifestReferenceMap("service", "../../test/templates/manifests", addOnConfig.References["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}
