package cmdhelpers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
)

const templatePath = "../../test/templates"

func TestPromptAddonValues(t *testing.T) {
	templateWriter := &writers.FileMapWriter{}
	dir, remove, err := setUpTempDir("helm")
	assert.Nil(t, err)

	ingressTemplate, err := handlers.GetTemplate("app-routing-ingress", "", dir, templateWriter)
	assert.Nil(t, err)
	assert.NotNil(t, ingressTemplate)

	ingressTemplate.Config.SetVariable("ingress-tls-cert-keyvault-uri", "test.keyvault.uri")
	ingressTemplate.Config.SetVariable("ingress-use-osm-mtls", "false")
	ingressTemplate.Config.SetVariable("ingress-host", "test.host")
	ingressTemplate.Config.SetVariable("service-name", "test")
	ingressTemplate.Config.SetVariable("service-namespace", "test")
	ingressTemplate.Config.SetVariable("service-port", "80")

	err = PromptAddonValues(dir, ingressTemplate.Config)
	assert.Nil(t, err)
	assert.Nil(t, remove())
}

func TestGetHelmReferenceMap(t *testing.T) {
	refMap := make(map[string]string)
	err := extractHelmValuesToMap("service", "../../test/templates/helm", ReferenceResources["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetKustomizeReferenceMap(t *testing.T) {
	refMap := make(map[string]string)
	err := extractKustomizeValuesToMap("service", "../../test/templates/kustomize", ReferenceResources["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func TestGetManifestReferenceMap(t *testing.T) {
	refMap := make(map[string]string)
	err := extractManifestValuesToMap("service", "../../test/templates/manifests", ReferenceResources["service"], refMap)
	assert.Nil(t, err)
	assert.NotEmpty(t, refMap)
}

func setUpTempDir(deploy string) (dir string, close func() error, err error) {
	templateWriter := &writers.LocalFSWriter{}
	draftconfig := &config.DraftConfig{}
	dir, err = ioutil.TempDir("", "addonTest")
	if err != nil {
		return
	}
	close = func() error {
		return os.RemoveAll(dir)
	}
	fs := os.DirFS(templatePath)
	if err = osutil.CopyDir(fs, deploy, dir, draftconfig, templateWriter); err != nil {
		return
	}

	return dir, close, err
}
