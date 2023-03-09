package workflows

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/osutil"
)

const (
	parentDirName      = "workflows"
	workflowFilePrefix = "azure-kubernetes-service"
	ymlExtension       = ".yml"
)

var (
	deployNameToWorkflow = map[string]*workflowType{
		"helm":      {deployPath: "/charts", workflowFileSuffix: "-helm"},
		"kustomize": {deployPath: "/base", workflowFileSuffix: "-kustomize"},
		"manifests": {deployPath: "/manifests"},
	}
)

type workflowType struct {
	deployPath         string
	workflowFileSuffix string
}

func CreateWorkflows(dest string, config *WorkflowConfig, workflows embed.FS) error {
	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return err
	}

	if err = updateProductionDeployments(deployType, dest, config); err != nil {
		return err
	}
	workflowType, ok := deployNameToWorkflow[deployType]
	if !ok {
		return errors.New("unsupported deployment type")
	}

	workflowTemplate := getWorkflowFile(workflowType, workflows)

	replaceWorkflowVars(deployType, config, workflowTemplate)

	ghWorkflowPath := dest + "/.github/workflows/"
	ghWorkflowFileName := ghWorkflowPath + workflowFilePrefix + workflowType.workflowFileSuffix + ymlExtension
	log.Debugf("writing workflow to %s", ghWorkflowPath)

	return writeWorkflow(ghWorkflowPath, ghWorkflowFileName, *workflowTemplate)
}

func updateProductionDeployments(deployType, dest string, config *WorkflowConfig) error {
	productionImage := fmt.Sprintf("%s.azurecr.io/%s", config.AcrName, config.ContainerName)
	switch deployType {
	case "helm":
		return setHelmContainerImage(dest+"/charts/production.yaml", productionImage)
	case "kustomize":
		return setDeploymentContainerImage(dest+"/overlays/production/deployment.yaml", productionImage)
	case "manifests":
		return setDeploymentContainerImage(dest+"/manifests/deployment.yaml", productionImage)
	}
	return nil
}

func replaceWorkflowVars(deployType string, config *WorkflowConfig, ghw *GitHubWorkflow) {
	envMap := make(map[string]string)
	envMap["AZURE_CONTAINER_REGISTRY"] = config.AcrName
	envMap["CONTAINER_NAME"] = config.ContainerName
	envMap["RESOURCE_GROUP"] = config.ResourceGroupName
	envMap["CLUSTER_NAME"] = config.AksClusterName

	switch deployType {
	case "helm":
		envMap["CHART_PATH"] = config.ChartsPath
		envMap["CHART_OVERRIDE_PATH"] = config.ChartsOverridePath

	case "manifests":
		envMap["DEPLOYMENT_MANIFEST_PATH"] = config.ManifestsPath

	case "kustomize":
		envMap["KUSTOMIZE_PATH"] = config.KustomizePath
	}

	ghw.Env = envMap

	ghw.On.Push.Branches[0] = config.BranchName
}

func setDeploymentContainerImage(filePath, productionImage string) error {

	decode := scheme.Codecs.UniversalDeserializer().Decode
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	k8sObj, _, err := decode(file, nil, nil)
	if err != nil {
		return err
	}
	deploy, ok := k8sObj.(*appsv1.Deployment)
	if !ok {
		return errors.New("could not decode kubernetes deployment")
	}

	if len(deploy.Spec.Template.Spec.Containers) != 1 {
		return errors.New("unsupported number of containers defined in the deployment spec")
	}

	deploy.Spec.Template.Spec.Containers[0].Image = productionImage

	printer := printers.YAMLPrinter{}

	out, err := os.OpenFile(filePath, os.O_RDWR, 0755)
	if err != nil {
		return nil
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Errorf("error closing file: %v", err)
		}
	}()

	return printer.PrintObj(deploy, out)
}

func setHelmContainerImage(filePath, productionImage string) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var deploy HelmProductionYaml
	err = yaml.Unmarshal(file, &deploy)
	if err != nil {
		return err
	}

	deploy.Image.Repository = productionImage

	out, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, out, 0644)
}

func getWorkflowFile(workflowType *workflowType, workflows embed.FS) *GitHubWorkflow {
	embedFilePath := path.Join(parentDirName, workflowFilePrefix+workflowType.workflowFileSuffix+ymlExtension)
	file, err := fs.ReadFile(workflows, embedFilePath)
	if err != nil {
		log.Fatal(err)
	}

	var ghw GitHubWorkflow

	err = yaml.Unmarshal(file, &ghw)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &ghw
}

func writeWorkflow(ghWorkflowPath, workflowFileName string, ghw GitHubWorkflow) error {
	workflowBytes, err := yaml.Marshal(ghw)
	if err != nil {
		return err
	}

	if err := osutil.EnsureDirectory(ghWorkflowPath); err != nil {
		return err
	}

	return os.WriteFile(workflowFileName, workflowBytes, 0644)
}
