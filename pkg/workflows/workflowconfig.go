package workflows

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
