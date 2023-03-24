package workflows

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Azure/draft/pkg/templatewriter/writers"
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

func createTempDeploymentFile(dirPath, fileName, path string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var source *os.File
	source, err = os.Open(path)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return err
	}
	return nil
}
func TestCreateWorkflows(t *testing.T) {
	dest := "."
	deployType := "helm"
	flagVariables := []string{}
	templatewriter := &writers.LocalFSWriter{}
	flagValuesMap := map[string]string{"AZURECONTAINERREGISTRY": "testAcr", "CONTAINERNAME": "testContainer", "RESOURCEGROUP": "testRG", "CLUSTERNAME": "testCluster", "BRANCHNAME": "testBranch", "BUILDCONTEXTPATH": "."}
	err := createTempDeploymentFile("charts", "charts/production.yaml", "../../test/templates/helm/charts/production.yaml")
	assert.Nil(t, err)
	assert.Nil(t, CreateWorkflows(dest, deployType, flagVariables, templatewriter, flagValuesMap))
	os.RemoveAll("charts")
	os.RemoveAll(".github")

	deployType = "kustomize"
	err = createTempDeploymentFile("overlays/production", "overlays/production/deployment.yaml", "../../test/templates/kustomize/overlays/production/deployment.yaml")
	assert.Nil(t, err)
	assert.Nil(t, CreateWorkflows(dest, deployType, flagVariables, templatewriter, flagValuesMap))
	os.RemoveAll("overlays")
	os.RemoveAll(".github")

	deployType = "manifests"
	err = createTempDeploymentFile("manifests", "manifests/deployment.yaml", "../../test/templates/manifests/manifests/deployment.yaml")
	assert.Nil(t, err)
	assert.Nil(t, CreateWorkflows(dest, deployType, flagVariables, templatewriter, flagValuesMap))
	os.RemoveAll("manifests")
	os.RemoveAll(".github")

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
