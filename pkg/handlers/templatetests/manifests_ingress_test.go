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
		{
			Name:            "valid ingress with app routing enabled and managed cert",
			TemplateName:    "ingress-manifests",
			FixturesBaseDir: "../../fixtures/manifests/ingress",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"INGRESSNAME":      "test-ingress",
				"PARTOF":           "test-app",
				"SERVICENAME":      "test-service",
				"SERVICEPORT":      "80",
				"ENABLEAPPROUTING": "true",
				"HASMANAGEDCERT":   "true",
				"CERTKEYVAULTURI":  "test.uri",
				"HOST":             "test-host.com",
				"PATH":             "/test/path",
				"INGRESSCLASSNAME": "webapprouting.kubernetes.azure.com",
			},
		},
		{
			Name:            "valid simple ingress",
			TemplateName:    "ingress-manifests",
			FixturesBaseDir: "../../fixtures/manifests/ingress",
			Version:         "0.0.1",
			Dest:            ".",
			TemplateWriter:  &writers.FileMapWriter{},
			VarMap: map[string]string{
				"INGRESSNAME":      "test-ingress",
				"PARTOF":           "test-app",
				"SERVICENAME":      "test-service",
				"SERVICEPORT":      "80",
				"HOST":             "test-host.com",
				"INGRESSCLASSNAME": "nginx",
			},
			FileNameOverride: map[string]string{
				"ingress.yaml": "ingress-simple.yaml",
			},
		},
	}

	for _, test := range tests {
		RunTemplateTest(t, test)
	}
}
