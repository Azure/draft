package templatetests

import (
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestManifestsIngressTemplates(t *testing.T) {
	tests := []TestInput{
		{
			Name:            "valid app-routing ingress",
			TemplateName:    "app-routing-ingress",
			FixturesBaseDir: "../../fixtures/addons/ingress",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"ingress-tls-cert-keyvault-uri": "test.uri",
				"ingress-use-osm-mtls":          "false",
				"ingress-host":                  "host",
				"service-name":                  "test-service",
				"service-namespace":             "test-namespace",
				"service-port":                  "80",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
