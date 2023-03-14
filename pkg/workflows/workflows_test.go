package workflows

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func createTempManifest(path string) (string, error) {
	file, err := ioutil.TempFile("", "*.yaml")
	if err != nil {
		return "", err
	}
	defer file.Close()

	var source *os.File
	source, err = os.Open(path)
	if err != nil {
		return "", err
	}
	defer source.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestUpdateProductionDeployments(t *testing.T) {
	flagValuesMap := map[string]string{"AZURECONTAINERREGISTRY": "testRegistry", "CONTAINERNAME": "testContainer"}
	testTemplateWriter := &writers.LocalFSWriter{}
	assert.Nil(t, updateProductionDeployments("", ".", flagValuesMap, testTemplateWriter))

	helmFileName, _ := createTempManifest("../../test/templates/helm_prod_values.yaml")
	deploymentFileName, _ := createTempManifest("../../test/templates/deployment.yaml")
	defer os.Remove(helmFileName)
	defer os.Remove(deploymentFileName)

	assert.Nil(t, setHelmContainerImage(helmFileName, "testImage", testTemplateWriter))

	helmDeploy := &HelmProductionYaml{}
	assert.Nil(t, helmDeploy.LoadFromFile(helmFileName))
	assert.Equal(t, "testImage", helmDeploy.Image.Repository)

	assert.Nil(t, setDeploymentContainerImage(deploymentFileName, "testImage"))
	decode := scheme.Codecs.UniversalDeserializer().Decode
	file, err := ioutil.ReadFile(deploymentFileName)
	assert.Nil(t, err)

	k8sObj, _, err := decode(file, nil, nil)
	assert.Nil(t, err)

	deploy, ok := k8sObj.(*appsv1.Deployment)
	assert.True(t, ok)
	assert.Equal(t, "testImage", deploy.Spec.Template.Spec.Containers[0].Image)
}
