package workflows

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/templatewriter/writers"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestUpdateProductionDeploymentsValid(t *testing.T) {
	testTemplateWriter := &writers.LocalFSWriter{}

	//test for valid helm deployment file
	helmFileName, _ := createTempManifest("../../test/templates/helm_prod_values.yaml")
	defer os.Remove(helmFileName)

	assert.Nil(t, setHelmContainerImage(helmFileName, "testImage", testTemplateWriter))

	helmDeploy := &HelmProductionYaml{}
	assert.Nil(t, helmDeploy.LoadFromFile(helmFileName))
	assert.Equal(t, "testImage", helmDeploy.Image.Repository)

	//test for valid deployment file
	deploymentFileName, _ := createTempManifest("../../test/templates/deployment.yaml")
	defer os.Remove(deploymentFileName)

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

func TestUpdateProductionDeploymentsInvalid(t *testing.T) {
	testTemplateWriter := &writers.LocalFSWriter{}

	//test for invalid helm deployment file
	tempFile, err := ioutil.TempFile("", "*.yaml")
	assert.Nil(t, err)
	defer os.Remove(tempFile.Name())
	yamlData := []byte(`not a valid yaml`)
	_, err = tempFile.Write(yamlData)
	assert.Nil(t, err)
	err = tempFile.Close()
	assert.Nil(t, err)
	assert.NotNil(t, setHelmContainerImage(tempFile.Name(), "testImage", testTemplateWriter))

	//test for invalid deployment file
	assert.NotNil(t, setDeploymentContainerImage(tempFile.Name(), "testImage"))

	//test for invalid k8sObj
	invalidDeploymentFile, _ := createTempManifest("../../test/templates/invalid_deployment.yaml")
	assert.Equal(t, errors.New("could not decode kubernetes deployment"), setDeploymentContainerImage(invalidDeploymentFile, "testImage"))

	//test for unsupported number of containers in the deployment spec
	invalidDeploymentFile, _ = createTempManifest("../../test/templates/unsupported_no_of_containers.yaml")
	defer os.Remove(invalidDeploymentFile)
	assert.Equal(t, errors.New("unsupported number of containers defined in the deployment spec"), setDeploymentContainerImage(invalidDeploymentFile, "testImage"))
}

func TestUpdateProductionDeploymentsMissing(t *testing.T) {
	draftConfig := &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "AZURECONTAINERREGISTRY",
				Value: "testRegistry",
			},
			{
				Name:  "CONTAINERNAME",
				Value: "testContainer",
			},
		},
	}
	testTemplateWriter := &writers.LocalFSWriter{}
	//test for missing deploy type
	assert.Nil(t, UpdateProductionDeployments("", ".", draftConfig, testTemplateWriter))

	//test for missing helm deployment file
	assert.NotNil(t, setHelmContainerImage("", "testImage", testTemplateWriter))

	//test for missing deployment file
	assert.NotNil(t, setDeploymentContainerImage("", "testImage"))
}

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
