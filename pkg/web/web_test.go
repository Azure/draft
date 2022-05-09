package web

import (
	"github.com/Azure/draft/pkg/types"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/osutil"
)

func createTempManifest(path string) (*os.File, error) {
	file, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var source *os.File
	source, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	defer source.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func TestAddAnnotationsKustomize(t *testing.T) {
	annotations := map[string]string{
		"kubernetes.azure.com/ingress-host":          "test.SA",
		"kubernetes.azure.com/tls-cert-keyvault-uri": "test.Cert",
	}

	annotatedManifest, fileErr := createTempManifest("../../test/templates/service_w_annotations.yaml")
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	defer os.Remove(annotatedManifest.Name())

	err := updateServiceAnnotationsForDeployment(annotatedManifest.Name(), "kustomize", annotations)
	assert.Nil(t, err)

	eKustomizeYaml := &types.ServiceYaml{}

	eKustomizeYaml.LoadFromFile(annotatedManifest.Name())

	assert.NotNil(t, eKustomizeYaml.Annotations)
	assert.Equal(t, annotations, eKustomizeYaml.Annotations)
}

func TestUpdateServiceFile(t *testing.T) {
	tempDest := "./../.."
	tempFile := tempDest + "/manifests/service.yaml"
	mockSa := &ServiceAnnotations{Host: "mockHost", Cert: "mockCert"}

	osutil.EnsureDirectory(tempDest + "/manifests")
	defer os.Remove(tempDest + "/manifests")
	osutil.EnsureFile(tempFile)
	defer os.Remove(tempFile)

	contents, err := ioutil.ReadFile("../../test/templates/service_w_annotations.yaml")
	assert.Nil(t, err)
	ioutil.WriteFile(tempFile, contents, 0644)

	err = UpdateServiceFile(mockSa, tempDest)
	assert.Nil(t, err)
	newContents, _ := ioutil.ReadFile(tempFile)

	assert.NotEqual(t, contents, newContents)
}

func TestAddAnnotationsHelm(t *testing.T) {
	annotations := map[string]string{
		"kubernetes.azure.com/ingress-host":          "test.SA",
		"kubernetes.azure.com/tls-cert-keyvault-uri": "test.Cert",
	}

	annotatedManifest, fileErr := createTempManifest("../../test/templates/helm_prod_values.yaml")
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	defer os.Remove(annotatedManifest.Name())

	deployNameToServiceYaml = map[string]*service{
		"kustomize": {file: annotatedManifest.Name()},
		"helm":      {file: annotatedManifest.Name()},
		"manifests": {file: annotatedManifest.Name()},
	}

	if err := updateServiceAnnotationsForDeployment(annotatedManifest.Name(), "helm", annotations); err != nil {
		t.Fatal(err)
	}

	eHelmYaml := &types.HelmProductionYaml{}
	eHelmYaml.LoadFromFile(annotatedManifest.Name())

	assert.NotNil(t, eHelmYaml.Service.Annotations)
	assert.Equal(t, annotations, eHelmYaml.Service.Annotations)
}

