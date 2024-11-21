package writers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/template"
)

func TestCopyDirToFileMap(t *testing.T) {

	templatewriter := &FileMapWriter{}
	err := osutil.CopyDir(template.Templates, "addons/azure/webapp_routing", "/test/dir", &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "ingress-tls-cert-keyvault-uri",
				Value: "https://test.vault.azure.net/secrets/test-secret",
			},
			{
				Name:  "ingress-use-osm-mtls",
				Value: "true",
			},
			{
				Name:  "ingress-host",
				Value: "testhost.com",
			},
		},
	}, templatewriter)
	assert.Nil(t, err)
	assert.NotNil(t, templatewriter.FileMap)
	assert.NotNil(t, templatewriter.FileMap["/test/dir/ingress.yaml"])
}
