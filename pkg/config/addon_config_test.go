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
