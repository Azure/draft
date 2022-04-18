package web

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestAddAnotationsKustomize(t *testing.T) {
	testSa := &ServiceAnnotations{}
	testSa.Host = "test host"
	testSa.Cert = "test cert"
	
	emptyManifest, fileErr := createTempManifest("../../test/templates/empty_service.yaml")
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	defer os.Remove(emptyManifest.Name())

	deployNameToServiceYaml = map[string]*service{
		"kustomize": {file: emptyManifest.Name(), annotation: "metadata.annotations"},
		"helm": {file: emptyManifest.Name(), annotation: "metadata.annotations"},
		"manifests": {file: emptyManifest.Name(), annotation: "metadata.annotations"},
	}

	if err := UpdateServiceFile(testSa, emptyManifest.Name()); err != nil {
		t.Fatal(err)
	}

	eManifestBytes, _ := os.ReadFile(emptyManifest.Name())
	annotatedManifest := string(eManifestBytes)
	oManifestBytes, _ := os.ReadFile("../../test/templates/empty_service.yaml.yaml")
	ogManifest := string(oManifestBytes)

	assert.False(t, annotatedManifest == ogManifest, "annotations weren't added")
}

func TestReplaceAnnotationsKustomize(t *testing.T) {
	testSa := &ServiceAnnotations{}
	testSa.Host = "test host"
	testSa.Cert = "test cert"

	annotatedManifest, fileErr := createTempManifest("../../test/templates/service_w_annotations.yaml")
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	defer os.Remove(annotatedManifest.Name())

	deployNameToServiceYaml = map[string]*service{
		"kustomize": {file: annotatedManifest.Name(), annotation: "metadata.annotations"},
		"helm": {file: annotatedManifest.Name(), annotation: "metadata.annotations"},
		"manifests": {file: annotatedManifest.Name(), annotation: "metadata.annotations"},
	}

	if err := UpdateServiceFile(testSa, annotatedManifest.Name()); err != nil {
		t.Fatal(err)
	}

	eManifestBytes, _ := os.ReadFile(annotatedManifest.Name())
	eManifest := string(eManifestBytes)
	ogManifestBytes, _ := os.ReadFile("../../test/templates/service_w_annotations.yaml")
	ogManifest := string(ogManifestBytes)

	assert.True(t, len(eManifest) == len(ogManifest), "annotations weren't replaced correctly")
}
