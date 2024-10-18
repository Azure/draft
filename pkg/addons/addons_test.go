package addons

import (
	"fmt"
	"github.com/Azure/draft/pkg/fixtures"
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
					Name:  "GENERATORLABEL",
					Value: "draft",
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

	// Validate generated content against the fixture
	generatedFilePath := fmt.Sprintf("%s/charts/templates/ingress.yaml", dir)
	generatedContent, err := os.ReadFile(generatedFilePath)
	assert.Nil(t, err)

	fixturePath := "../fixtures/addons/helm/ingress.yaml"
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Fatalf("Fixture file does not exist at path: %s", fixturePath)
	}

	err = fixtures.ValidateContentAgainstFixture(generatedContent, fixturePath)
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
					Name:  "GENERATORLABEL",
					Value: "draft",
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

	// Validate generated content against the fixture
	generatedFilePath := fmt.Sprintf("%s/overlays/production/ingress.yaml", dir)
	generatedContent, err := os.ReadFile(generatedFilePath)
	assert.Nil(t, err)

	fixturePath := "../fixtures/addons/kustomize/ingress.yaml"
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Fatalf("Fixture file does not exist at path: %s", fixturePath)
	}

	err = fixtures.ValidateContentAgainstFixture(generatedContent, fixturePath)
	assert.Nil(t, err)

	assert.Nil(t, remove())
}

func TestNilDraftConfig(t *testing.T) {
	templateWriter := &writers.LocalFSWriter{}
	addonConfig := AddonConfig{}
	dir, remove, err := setUpTempDir("helm")
	assert.Nil(t, err)

	err = GenerateAddon(template.Addons, "azure", "webapp_routing", dir, addonConfig, templateWriter)
	assert.NotNil(t, err, "should fail with nil DraftConfig")
	assert.Equal(t, "DraftConfig is nil", err.Error())

	err = PromptAddonValues(dir, &addonConfig)
	assert.NotNil(t, err, "should fail with nil DraftConfig")
	assert.Equal(t, "draftConfig is nil", err.Error())

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
