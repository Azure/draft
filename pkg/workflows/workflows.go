package workflows

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/templatewriter"
)

//go:generate cp -r ../../starterWorkflows ./workflows

var (
	//go:embed workflows
	workflows     embed.FS
	parentDirName = "workflows"

	workflowFilePrefix   = "azure-kubernetes-service"
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

func CreateWorkflows(dest string, config *WorkflowConfig, flagVariables []string, templateWriter templatewriter.TemplateWriter) error {
	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return err
	}

	if err = updateProductionDeployments(deployType, dest, config, templateWriter); err != nil {
		return err
	}
	workflow, ok := deployNameToWorkflow[deployType]
	if !ok {
		return errors.New("unsupported deployment type")
	}

	workflowTemplate := getWorkflowFile(workflow)

	if err = replaceWorkflowVars(deployType, config, workflowTemplate, flagVariables); err != nil {
		return err
	}

	ghWorkflowPath := dest + "/.github/workflows/"
	ghWorkflowFileName := ghWorkflowPath + workflowFilePrefix + workflow.workflowFileSuffix + ".yml"
	log.Debugf("writing workflow to %s", ghWorkflowPath)

	return writeWorkflow(ghWorkflowPath, ghWorkflowFileName, *workflowTemplate, templateWriter)
}

func updateProductionDeployments(deployType, dest string, config *WorkflowConfig, templateWriter templatewriter.TemplateWriter) error {
	productionImage := fmt.Sprintf("%s.azurecr.io/%s", config.AcrName, config.ContainerName)
	switch deployType {
	case "helm":
		return setHelmContainerImage(dest+"/charts/production.yaml", productionImage, templateWriter)
	case "kustomize":
		return setDeploymentContainerImage(dest+"/overlays/production/deployment.yaml", productionImage)
	case "manifests":
		return setDeploymentContainerImage(dest+"/manifests/deployment.yaml", productionImage)
	}
	return nil
}

func replaceWorkflowVars(deployType string, config *WorkflowConfig, ghw *GitHubWorkflow, flagVariables []string) error {
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

	for _, flagVar := range flagVariables {
		flagVarName, flagVarValue, ok := strings.Cut(flagVar, "=")
		if !ok {
			return fmt.Errorf("invalid variable format: %s", flagVar)
		}
		envMap[flagVarName] = flagVarValue
		log.Debugf("flag variable %s=%s", flagVarName, flagVarValue)
	}

	ghw.Env = envMap

	ghw.On.Push.Branches[0] = config.BranchName

	return nil
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

func setHelmContainerImage(filePath, productionImage string, templateWriter templatewriter.TemplateWriter) error {
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

	return templateWriter.WriteFile(filePath, out)
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

func writeWorkflow(ghWorkflowPath, workflowFileName string, ghw GitHubWorkflow, templateWriter templatewriter.TemplateWriter) error {
	workflowBytes, err := yaml.Marshal(ghw)
	if err != nil {
		return err
	}

	if err := templateWriter.EnsureDirectory(ghWorkflowPath); err != nil {
		return err
	}

	return templateWriter.WriteFile(workflowFileName, workflowBytes)
}
