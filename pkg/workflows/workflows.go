package workflows

import (
	"embed"
	"errors"
	"fmt"
	"github.com/Azure/draftv2/pkg/filematches"
	"github.com/Azure/draftv2/pkg/osutil"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
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
	productionImage := fmt.Sprintf("%s.azurecr.io/%s:latest", config.AcrName, config.ContainerName)
	switch deployType {
	case "helm":
		return openAndReplace(dest+"/charts/production.yaml", productionImage)
	case "kustomize":
		return openAndReplace(dest+"/overlays/production/deployment.yaml", productionImage)
	}
	return nil
}

func openAndReplace(filePath, productionImage string) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	file = []byte(strings.ReplaceAll(string(file), productionImagePlaceholder, productionImage))

	return ioutil.WriteFile(filePath, file, 0644)
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
}

func removeStep(steps []map[string]interface{}, index int) []map[string]interface{} {
	return append(steps[:index], steps[index+1:]...)
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

	marshaledYaml := string(workflowBytes)
	log.Debug(marshaledYaml)
	if err := osutil.EnsureDirectory(ghWorkflowPath); err != nil {
		return err
	}

	return os.WriteFile(workflowFileName, workflowBytes, 0644)
}
