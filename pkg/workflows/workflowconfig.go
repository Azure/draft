package workflows

import (
	"github.com/Azure/draft/pkg/prompts"
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
	BuildContextPath   string
}

func (config *WorkflowConfig) SetFlagValuesToMap() map[string]string {
	flagValuesMap := make(map[string]string)
	if config.AcrName != "" {
		flagValuesMap["AZURECONTAINERREGISTRY"] = config.AcrName
	}

	if config.ContainerName != "" {
		flagValuesMap["CONTAINERNAME"] = config.ContainerName
	}

	if config.ResourceGroupName != "" {
		flagValuesMap["RESOURCEGROUP"] = config.ResourceGroupName
	}

	if config.AksClusterName != "" {
		flagValuesMap["CLUSTERNAME"] = config.AksClusterName
	}

	if config.BranchName != "" {
		flagValuesMap["BRANCHNAME"] = config.BranchName
	}

	if config.BuildContextPath == "" {
		config.BuildContextPath = prompts.GetInputFromPrompt("path to the docker build context, usually .")
	}

	config.ChartsPath = "./charts"
	config.ChartsOverridePath = "./charts/production.yaml"
	config.ManifestsPath = "./manifests"
	config.KustomizePath = "./overlays/production"
}
