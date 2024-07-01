package addons

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

const templatePath = "../../test/templates"

func TestGenerateAddonErrors(t *testing.T) {
	var addonConfig AddonConfig
	templateWriter := &writers.LocalFSWriter{}
	err := GenerateAddon(template.Addons, "azure", "webapp_routing", "fakeDest", addonConfig, templateWriter)
	assert.NotNil(t, err, "should fail with fake destination")

	err = GenerateAddon(template.Addons, "azure", "fakeAddon", "../../test/templates/helm", addonConfig, templateWriter)
	assert.NotNil(t, err, "should fail with fake addon name")

	err = GenerateAddon(template.Addons, "fakeProvider", "fakeAddon", "../../test/templates/helm", addonConfig, templateWriter)
	assert.NotNil(t, err, "should fail with fake provider name")
}

func TestGenerateHelmAddonSuccess(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	addonConfig := AddonConfig{
		DraftConfig: &config.DraftConfig{
			Variables: []*config.BuilderVar{
				{
					Name:  "ingress-tls-cert-keyvault-uri",
					Value: "test.uri",
				},
				{
					Name:  "ingress-use-osm-mtls",
					Value: "false",
				},
				{
					Name:  "ingress-host",
					Value: "host",
				},
				{
					Name:  "service-namespace",
					Value: "test-namespace",
				},
				{
					Name:  "service-name",
					Value: "test-service",
				},
				{
					Name:  "service-port",
					Value: "80",
				},
			},
		},
	}
	dir, remove, err := setUpTempDir("helm")
	assert.Nil(t, err)

	err = GenerateAddon(template.Addons, "azure", "webapp_routing", dir, addonConfig, templateWriter)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func TestGenerateKustomizeAddonSuccess(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	addonConfig := AddonConfig{
		DraftConfig: &config.DraftConfig{
			Variables: []*config.BuilderVar{
				{
					Name:  "ingress-tls-cert-keyvault-uri",
					Value: "test.uri",
				},
				{
					Name:  "ingress-use-osm-mtls",
					Value: "false",
				},
				{
					Name:  "ingress-host",
					Value: "host",
				},
				{
					Name:  "service-namespace",
					Value: "test-namespace",
				},
				{
					Name:  "service-name",
					Value: "test-service",
				},
				{
					Name:  "service-port",
					Value: "80",
				},
			},
		},
	}
	dir, remove, err := setUpTempDir("kustomize")
	assert.Nil(t, err)

	err = GenerateAddon(template.Addons, "azure", "webapp_routing", dir, addonConfig, templateWriter)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func setUpTempDir(deploy string) (dir string, close func() error, err error) {
	templateWriter := &writers.LocalFSWriter{}
	addonConfig := AddonConfig{
		DraftConfig: &config.DraftConfig{},
	}
	dir, err = ioutil.TempDir("", "addonTest")
	if err != nil {
		return
	}
	close = func() error {
		return os.RemoveAll(dir)
	}
	fs := os.DirFS(templatePath)
	if err = osutil.CopyDir(fs, deploy, dir, addonConfig.DraftConfig, templateWriter); err != nil {
		return
	}

	return dir, close, err
}
