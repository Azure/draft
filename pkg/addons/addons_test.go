package addons

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

const templatePath = "../../test/templates"

func TestGenerateAddonErrors(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	userInputs := map[string]string{
		"test": "test",
	}
	err := GenerateAddon(template.Addons, "azure", "webapp_routing", "fakeDest", userInputs, templateWriter)
	assert.NotNil(t, err, "should fail with fake destination")

	err = GenerateAddon(template.Addons, "azure", "fakeAddon", "../../test/templates/helm", userInputs, templateWriter)
	assert.NotNil(t, err, "should fail with fake addon name")

	err = GenerateAddon(template.Addons, "fakeProvider", "fakeAddon", "../../test/templates/helm", userInputs, templateWriter)
	assert.NotNil(t, err, "should fail with fake provider name")
}

func TestGenerateHelmAddonSuccess(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	var correctUserInputs = map[string]string{
		"ingress-tls-cert-keyvault-uri": "test.uri",
		"ingress-use-osm-mtls":          "false",
		"ingress-host":                  "host",
		"service-namespace":             "test-namespace",
		"service-name":                  "test-service",
		"service-port":                  "80",
	}
	dir, remove, err := setUpTempDir("helm")
	assert.Nil(t, err)

	err = GenerateAddon(template.Addons, "azure", "webapp_routing", dir, correctUserInputs, templateWriter)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func TestGenerateKustomizeAddonSuccess(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	var correctUserInputs = map[string]string{
		"ingress-tls-cert-keyvault-uri": "test.uri",
		"ingress-use-osm-mtls":          "false",
		"ingress-host":                  "host",
		"service-namespace":             "test-namespace",
		"service-name":                  "test-service",
		"service-port":                  "80",
	}
	dir, remove, err := setUpTempDir("kustomize")
	assert.Nil(t, err)

	err = GenerateAddon(template.Addons, "azure", "webapp_routing", dir, correctUserInputs, templateWriter)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func setUpTempDir(deploy string) (dir string, close func() error, err error) {
	templateWriter := &writers.LocalFSWriter{}
	dir, err = ioutil.TempDir("", "addonTest")
	if err != nil {
		return
	}
	close = func() error {
		return os.RemoveAll(dir)
	}
	fs := os.DirFS(templatePath)
	if err = osutil.CopyDir(fs, deploy, dir, nil, nil, templateWriter); err != nil {
		return
	}

	return dir, close, err
}
