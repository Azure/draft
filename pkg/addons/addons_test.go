package addons

import (
	"github.com/Azure/draft/pkg/osutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

const templatePath = "../../test/templates"

func TestGenerateAddonErrors(t *testing.T) {
	userInputs := map[string]string{
		"test": "test",
	}
	err := GenerateAddon("azure", "webapp_routing", "fakeDest", userInputs)
	assert.NotNil(t, err, "should fail with fake destination")

	err = GenerateAddon("azure", "fakeAddon", "../../test/templates/helm", userInputs)
	assert.NotNil(t, err, "should fail with fake addon name")

	err = GenerateAddon("fakeProvider", "fakeAddon", "../../test/templates/helm", userInputs)
	assert.NotNil(t, err, "should fail with fake provider name")
}

func TestGenerateHelmAddonSuccess(t *testing.T) {
	var correctUserInputs = map[string]string{
		"ingress-tls-cert-keyvault-uri": "test.uri",
		"ingress-use-osm-mtls":          "false",
		"ingress-host":                  "host",
	}
	dir, remove, err := setUpTempDir("helm")
	assert.Nil(t, err)

	err = GenerateAddon("azure", "webapp_routing", dir, correctUserInputs)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func TestGenerateKustomizeAddonSuccess(t *testing.T) {
	var correctUserInputs = map[string]string{
		"ingress-tls-cert-keyvault-uri": "test.uri",
		"ingress-use-osm-mtls":          "false",
		"ingress-host":                  "host",
	}
	dir, remove, err := setUpTempDir("kustomize")
	assert.Nil(t, err)

	err = GenerateAddon("azure", "webapp_routing", dir, correctUserInputs)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func setUpTempDir(deploy string) (dir string, close func() error, err error) {
	dir, err = ioutil.TempDir("", "addonTest")
	if err != nil {
		return
	}
	close = func() error {
		return os.RemoveAll(dir)
	}
	fs := os.DirFS(templatePath)
	if err = osutil.CopyDir(fs, deploy, dir, nil, nil); err != nil {
		return
	}

	return dir, close, err
}
