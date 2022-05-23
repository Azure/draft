package config

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/Azure/draft/template"
)

func TestGetHelmReferenceMap(t *testing.T) {
	configBytes, err := fs.ReadFile(template.Addons, "addons/azure/webapp_routing/draft.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	assert.Nil(t, err)

	refMap := make(map[string]string)
	err = extractHelmValuesToMap("service", "../../test/templates/helm", addOnConfig.ReferenceComponents["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetKustomizeReferenceMap(t *testing.T) {
	configBytes, err := fs.ReadFile(template.Addons, "addons/azure/webapp_routing/draft.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	refMap := make(map[string]string)

	err = extractKustomizeValuesToMap("service", "../../test/templates/kustomize", addOnConfig.ReferenceComponents["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetManifestReferenceMap(t *testing.T) {
	configBytes, err := fs.ReadFile(template.Addons, "addons/azure/webapp_routing/draft.yaml")
	assert.Nil(t, err)

	var addOnConfig AddonConfig
	err = yaml.Unmarshal(configBytes, &addOnConfig)
	refMap := make(map[string]string)

	err = extractManifestValuesToMap("service", "../../test/templates/manifests", addOnConfig.ReferenceComponents["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}
