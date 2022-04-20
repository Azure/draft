package workflows

import (
	"embed"
	"errors"
	"fmt"
	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/osutil"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
)

//go:generate cp -r ../../starterWorkflows ./workflows

var (
	//go:embed workflows
	workflows     embed.FS
	parentDirName = "workflows"

	workflowFilePrefix         = "azure-kubernetes-service"
	productionImagePlaceholder = "{{PRODUCTION_CONTAINER_IMAGE}}"
	deployNameToWorkflow       = map[string]*workflowType{
		"helm":      {deployPath: "/charts", workflowFileSuffix: "-helm"},
		"kustomize": {deployPath: "/base", workflowFileSuffix: "-kustomize"},
		"manifests": {deployPath: "/manifests"},
	}
)

type workflowType struct {
	deployPath         string
	workflowFileSuffix string
}

func CreateWorkflows(dest string, config *WorkflowConfig) error {
	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return err
	}

	if err = updateProductionDeployments(deployType, dest, config); err != nil {
		return err
	}
	workflow, ok := deployNameToWorkflow[deployType]
	if !ok {
		return errors.New("unsupported deployment type")
	}

	workflowTemplate := getWorkflowFile(workflow)

	replaceWorkflowVars(deployType, config, workflowTemplate)

	ghWorkflowPath := dest + "/.github/workflows/"
	ghWorkflowFileName := ghWorkflowPath + workflowFilePrefix + workflow.workflowFileSuffix + ".yml"
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
		envMap["CHART_PATH"] = config.chartsPath
		envMap["CHART_OVERRIDE_PATH"] = config.chartsOverridePath

	case "manifests":
		envMap["DEPLOYMENT_MANIFEST_PATH"] = config.manifestsPath

	case "kustomize":
		envMap["KUSTOMIZE_PATH"] = config.kustomizePath
	}

	ghw.Env = envMap
	editedJob, ok := ghw.Jobs["build"]
	if ok {
		editedJob.Steps = removeStep(editedJob.Steps, 4)
		ghw.Jobs["build"] = editedJob
	}

	ghw.On.Push.Branches[0] = config.BranchName
}

func removeStep(steps []map[string]interface{}, index int) []map[string]interface{} {
	return append(steps[:index], steps[index+1:]...)
}

func setDeploymentContainerImage(filePath, productionImage string) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var deploy DeploymentYaml
	err = yaml.Unmarshal(file, &deploy)
	if err != nil {
		return err
	}

	deploy.Spec.Template.Spec.Containers[0].Image = productionImage

	out, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, out, 0644)
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

	deploy.ImageKey.Repository = productionImage

	out, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, out, 0644)
}

func getWorkflowFile(workflow *workflowType) *GitHubWorkflow {
	embedFilePath := parentDirName + "/" + workflowFilePrefix + workflow.workflowFileSuffix + ".yml"

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
