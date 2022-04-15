package workflows

import (
	"github.com/Azure/draftv2/pkg/prompts"
)

//GitHubWorkflow is a rough struct to allow for yaml editing including deletion of Job steps
type GitHubWorkflow struct {
	Name string
	On   map[string]interface{}
	Env  map[string]string
	Jobs map[string]job
}

type job struct {
	Permissions map[string]string
	RunsOn      string `yaml:"runs-on"`
	Steps       []map[string]interface{}
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
	config.chartsOverridePath = "./charts/production.yaml"
	config.manifestsPath = "./manifests"
	config.kustomizePath = "./overlays/production"
}
