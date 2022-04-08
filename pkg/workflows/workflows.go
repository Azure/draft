package workflows

import (
	"embed"
	"errors"
	"github.com/Azure/draftv2/pkg/filematches"
	"github.com/Azure/draftv2/pkg/osutil"
	"github.com/Azure/draftv2/pkg/prompts"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"strings"
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

type WorkflowConfig struct {
	AcrName            string
	ContainerName      string
	ResourceGroupName  string
	AksClusterName     string
	manifestsPath      string
	chartsPath         string
	chartsOverridePath string
	kustomizePath      string
}

func CreateWorkflows(dest string, config *WorkflowConfig) error {
	deployType, err := filematches.FindDraftDeploymentFiles(dest)
	if err != nil {
		return err
	}

	workflow, ok := deployNameToWorkflow[deployType]
	if !ok {
		return errors.New("unsupported deployment type")
	}

	workflowTemplate := getWorkflowFile(workflow)

	workflowTemplate = replaceWorkflowVars(workflowTemplate, config)

	ghWorkflowPath := dest + "/.github/workflows/"
	log.Debugf("writing workflow to %s", ghWorkflowPath)

	if err := osutil.EnsureDirectory(ghWorkflowPath); err != nil {
		return err
	}

	if err := os.WriteFile(
		ghWorkflowPath+workflowFilePrefix+workflow.workflowFileSuffix+".yml",
		[]byte(workflowTemplate),
		0644); err != nil {
		return err
	}

	return nil
}

func replaceWorkflowVars(workflowTemplate string, config *WorkflowConfig) string {
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-azure-container-registry", config.AcrName)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-container-name", config.ContainerName)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-resource-group", config.ResourceGroupName)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-cluster-name", config.AksClusterName)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-chart-path", config.chartsPath)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-chart-override-path", config.chartsOverridePath)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-deployment-manifest-path", config.manifestsPath)
	workflowTemplate = strings.ReplaceAll(workflowTemplate, "your-kustomize-path", config.kustomizePath)
	return workflowTemplate
}

func (config *WorkflowConfig) ValidateAndFillConfig() {
	if config.AcrName == "" {
		config.AcrName = prompts.GetInputFromPrompt("container registry name")
	}

	if config.ContainerName == "" {
		config.ContainerName = prompts.GetInputFromPrompt("container name")
	}

	if config.ResourceGroupName == "" {
		config.ResourceGroupName = prompts.GetInputFromPrompt("cluster resource group name")
	}

	if config.AksClusterName == "" {
		config.AksClusterName = prompts.GetInputFromPrompt("AKS cluster name")
	}

	config.chartsPath = "./charts"
	config.chartsOverridePath = "./charts/values.yaml"
	config.manifestsPath = "./manifests"
	config.kustomizePath = "./base"
}

func getWorkflowFile(workflow *workflowType) string {
	embedFilePath := parentDirName + "/" + workflowFilePrefix + workflow.workflowFileSuffix + ".yml"

	file, err := fs.ReadFile(workflows, embedFilePath)
	if err != nil {
		log.Fatal(err)
	}

	return string(file)
}
