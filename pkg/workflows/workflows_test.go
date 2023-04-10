package workflows

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

func TestCreateWorkflows(t *testing.T) {
	dest := "."
	deployType := "helm"
	flagVariables := []string{}
	templatewriter := &writers.LocalFSWriter{}
	flagValuesMap := map[string]string{"AZURECONTAINERREGISTRY": "testAcr", "CONTAINERNAME": "testContainer", "RESOURCEGROUP": "testRG", "CLUSTERNAME": "testCluster", "BRANCHNAME": "testBranch"}
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
	//test for missing deployment file path
	assert.NotNil(t, CreateWorkflows(dest, deployType, flagVariables, templatewriter, flagValuesMap))

	//test for invalid deployType
	deployType = "testInvalidDeployType"
	assert.NotNil(t, CreateWorkflows(dest, deployType, flagVariables, templatewriter, flagValuesMap))
}

func TestUpdateProductionDeployments(t *testing.T) {
	flagValuesMap := map[string]string{"AZURECONTAINERREGISTRY": "testRegistry", "CONTAINERNAME": "testContainer"}
	testTemplateWriter := &writers.LocalFSWriter{}
	//test for missing deploy type
	assert.Nil(t, updateProductionDeployments("", ".", flagValuesMap, testTemplateWriter))

	//test for missing helm deployment file
	assert.NotNil(t, setHelmContainerImage("", "testImage", testTemplateWriter))

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

	//test for valid helm deployment file
	helmFileName, _ := createTempManifest("../../test/templates/helm_prod_values.yaml")
	defer os.Remove(helmFileName)

	assert.Nil(t, setHelmContainerImage(helmFileName, "testImage", testTemplateWriter))

	helmDeploy := &HelmProductionYaml{}
	assert.Nil(t, helmDeploy.LoadFromFile(helmFileName))
	assert.Equal(t, "testImage", helmDeploy.Image.Repository)

	//test for missing deployment file
	assert.NotNil(t, setDeploymentContainerImage("", "testImage"))

	//test for invalid deployment file
	assert.NotNil(t, setDeploymentContainerImage(tempFile.Name(), "testImage"))

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

	//test for invalid k8sObj
	invalidDeploymentFile, _ := createTempManifest("../../test/templates/invalid_deployment.yaml")
	assert.Equal(t, errors.New("could not decode kubernetes deployment"), setDeploymentContainerImage(invalidDeploymentFile, "testImage"))

	//test for unsupported number of containers in the deployment spec
	invalidDeploymentFile, _ = createTempManifest("../../test/templates/unsupported_no_of_containers.yaml")
	defer os.Remove(invalidDeploymentFile)
	assert.Equal(t, errors.New("unsupported number of containers defined in the deployment spec"), setDeploymentContainerImage(invalidDeploymentFile, "testImage"))
}

func TestLoadConfig(t *testing.T) {
	fakeFS, err := createMockWorkflowTemplatesFS()
	assert.Nil(t, err)
	w, err := createMockWorkflow("workflows", fakeFS)
	assert.Nil(t, err)

	cases := []loadConfTestCase{
		{"helm", true},
		{"kustomize", true},
		{"manifests", true},
		{"fake", false},
		{"emptyDir", false},
		{"corrupted", false},
	}

	for _, c := range cases {
		if c.isNil {
			_, err = w.loadConfig(c.deployType)
			assert.Nil(t, err)
		} else {
			_, err = w.loadConfig(c.deployType)
			assert.NotNil(t, err)
		}
	}
}

func TestPopulateConfigs(t *testing.T) {
	fakeFS, err := createMockWorkflowTemplatesFS()
	assert.Nil(t, err)

	w, err := createMockWorkflow("workflows", fakeFS)
	assert.Nil(t, err)

	w.populateConfigs()
	assert.Equal(t, 5, len(w.configs)) // includes emptyDir and corrupted so 2 additional configs

	w, err = createTestWorkflowEmbed("workflows")
	assert.Nil(t, err)

	w.populateConfigs()
	assert.Equal(t, 3, len(w.configs))
}

func TestCreateWorkflowFiles(t *testing.T) {
	templatewriter := &writers.LocalFSWriter{}
	customInputs := map[string]string{"AZURECONTAINERREGISTRY": "testAcr", "CONTAINERNAME": "testContainer", "RESOURCEGROUP": "testRG", "CLUSTERNAME": "testCluster", "BRANCHNAME": "testBranch", "CHARTPATH": "testPath", "CHARTOVERRIDEPATH": "testOverridePath"}
	badInputs := map[string]string{}

	workflowTemplate, err := createMockWorkflowTemplatesFS()
	assert.Nil(t, err)

	mockWF, err := createMockWorkflow("workflows", workflowTemplate)
	assert.Nil(t, err)

	mockWF.populateConfigs()

	err = mockWF.createWorkflowFiles("fakeDeployType", customInputs, templatewriter)
	assert.NotNil(t, err)

	err = mockWF.createWorkflowFiles("helm", customInputs, templatewriter)
	assert.Nil(t, err)

	err = mockWF.createWorkflowFiles("helm", badInputs, templatewriter)
	assert.NotNil(t, err)

}

type loadConfTestCase struct {
	deployType string
	isNil      bool
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

// Creates a copy of the embeded files in memory and returns a passable fs to use for testing
func createMockWorkflowTemplatesFS() (fs.FS, error) {
	rootPath := "workflows/"
	embedFiles, err := embedutils.EmbedFStoMapWithFiles(template.Workflows, "workflows")
	if err != nil {
		return nil, fmt.Errorf("failed to readDir: %w in embeded files", err)
	}

	mockFS := fstest.MapFS{}

	for path, file := range embedFiles {
		if file.IsDir() {
			mockFS[rootPath+path] = &fstest.MapFile{Mode: fs.ModeDir}
		} else {
			bytes, err := template.Workflows.ReadFile(rootPath + path)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}
			mockFS[rootPath+path] = &fstest.MapFile{Data: bytes}
		}
	}

	mockFS[rootPath+"emptyDir"] = &fstest.MapFile{Mode: fs.ModeDir}
	mockFS[rootPath+"corrupted"] = &fstest.MapFile{Mode: fs.ModeDir}
	mockFS[rootPath+"corrupted/draft.yaml"] = &fstest.MapFile{Data: []byte("fake yaml data")}

	return mockFS, nil
}

func createMockWorkflow(dirPath string, mockWorkflowTemplates fs.FS) (*Workflows, error) {
	dest := "."

	deployMap, err := fsToMap(mockWorkflowTemplates, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed fsToMap: %w", err)
	}

	w := &Workflows{
		workflows:         deployMap,
		dest:              dest,
		configs:           make(map[string]*config.DraftConfig),
		workflowTemplates: mockWorkflowTemplates,
	}

	return w, nil
}

func createTestWorkflowEmbed(dirPath string) (*Workflows, error) {
	dest := "."

	deployMap, err := embedutils.EmbedFStoMap(template.Workflows, "workflows")
	if err != nil {
		return nil, fmt.Errorf("failed to create deployMap: %w", err)
	}

	w := &Workflows{
		workflows:         deployMap,
		dest:              dest,
		configs:           make(map[string]*config.DraftConfig),
		workflowTemplates: template.Workflows,
	}

	return w, nil
}

func fsToMap(fsFS fs.FS, path string) (map[string]fs.DirEntry, error) {
	files, err := fs.ReadDir(fsFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadDir: %w", err)
	}

	mapping := make(map[string]fs.DirEntry)

	for _, f := range files {
		if f.IsDir() {
			mapping[f.Name()] = f
		}
	}

	return mapping, nil
}
