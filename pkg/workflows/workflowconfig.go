package workflows

import "github.com/Azure/draft/pkg/config"

type WorkflowConfig struct {
	WorkflowName         string
	BranchName           string
	AcrResourceGroup     string
	AcrName              string
	ContainerName        string
	ClusterResourceGroup string
	ClusterName          string
	Dockerfile           string
	BuildContextPath     string
	Namespace            string
	PrivateCluster       string
}

func (wc *WorkflowConfig) SetFlagsToValues(draftConfig *config.DraftConfig, varIdxMap map[string]int) {
	draftConfig.Variables[varIdxMap["WORKFLOWNAME"]].Value = wc.WorkflowName
	draftConfig.Variables[varIdxMap["BRANCHNAME"]].Value = wc.BranchName
	draftConfig.Variables[varIdxMap["ACRRESOURCEGROUP"]].Value = wc.AcrResourceGroup
	draftConfig.Variables[varIdxMap["AZURECONTAINERREGISTRY"]].Value = wc.AcrName
	draftConfig.Variables[varIdxMap["CONTAINERNAME"]].Value = wc.ContainerName
	draftConfig.Variables[varIdxMap["CLUSTERRESOURCEGROUP"]].Value = wc.ClusterResourceGroup
	draftConfig.Variables[varIdxMap["CLUSTERNAME"]].Value = wc.ClusterName
	draftConfig.Variables[varIdxMap["DOCKERFILE"]].Value = wc.Dockerfile
	draftConfig.Variables[varIdxMap["BUILDCONTEXTPATH"]].Value = wc.BuildContextPath
	draftConfig.Variables[varIdxMap["NAMESPACE"]].Value = wc.Namespace
	draftConfig.Variables[varIdxMap["PRIVATECLUSTER"]].Value = wc.PrivateCluster
}

func (config *WorkflowConfig) SetFlagValuesToMap() map[string]string {
	flagValuesMap := make(map[string]string)
	if config.WorkflowName != "" {
		flagValuesMap["WORKFLOWNAME"] = config.WorkflowName
	}

	if config.BranchName != "" {
		flagValuesMap["BRANCHNAME"] = config.BranchName
	}

	if config.AcrResourceGroup != "" {
		flagValuesMap["ACRRESOURCEGROUP"] = config.AcrResourceGroup
	}

	if config.AcrName != "" {
		flagValuesMap["AZURECONTAINERREGISTRY"] = config.AcrName
	}

	if config.ContainerName != "" {
		flagValuesMap["CONTAINERNAME"] = config.ContainerName
	}

	if config.ClusterResourceGroup != "" {
		flagValuesMap["CLUSTERRESOURCEGROUP"] = config.ClusterResourceGroup
	}

	if config.ClusterName != "" {
		flagValuesMap["CLUSTERNAME"] = config.ClusterName
	}

	if config.Dockerfile != "" {
		flagValuesMap["DOCKERFILE"] = config.Dockerfile
	}

	if config.BuildContextPath != "" {
		flagValuesMap["BUILDCONTEXTPATH"] = config.BuildContextPath
	}

	if config.Namespace != "" {
		flagValuesMap["NAMESPACE"] = config.Namespace
	}

	if config.PrivateCluster != "" {
		flagValuesMap["PRIVATECLUSTER"] = config.PrivateCluster
	}

	return flagValuesMap
}
