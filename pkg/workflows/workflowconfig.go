package workflows

import (
	"github.com/Azure/draft/pkg/prompts"
	"strings"
)

type WorkflowConfig struct {
	AcrName            string
	ContainerName      string
	ResourceGroupName  string
	AksClusterName     string
	BranchName         string
	ManifestsPath      string
	ChartsPath         string
	ChartsOverridePath string
	KustomizePath      string
}

func (config *WorkflowConfig) ValidateAndFillConfig() {
	if config.AcrName == "" {
		config.AcrName = strings.ToLower(prompts.GetInputFromPrompt("container registry name"))
	}

	if config.ContainerName == "" {
		config.ContainerName = strings.ToLower(prompts.GetInputFromPrompt("container name"))
	}

	if config.ResourceGroupName == "" {
		config.ResourceGroupName = prompts.GetInputFromPrompt("cluster resource group name")
	}

	if config.AksClusterName == "" {
		config.AksClusterName = prompts.GetInputFromPrompt("AKS cluster name")
	}

	if config.BranchName == "" {
		config.BranchName = prompts.GetInputFromPrompt("name of the repository branch to deploy from, usually main")
	}

	config.ChartsPath = "./charts"
	config.ChartsOverridePath = "./charts/production.yaml"
	config.ManifestsPath = "./manifests"
	config.KustomizePath = "./overlays/production"
}
