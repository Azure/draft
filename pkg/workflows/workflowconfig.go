package workflows

type WorkflowConfig struct {
	AcrName           string
	ContainerName     string
	ResourceGroupName string
	AksClusterName    string
	BranchName        string
	BuildContextPath  string
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

	if config.BuildContextPath != "" {
		flagValuesMap["BUILDCONTEXTPATH"] = config.BuildContextPath
	}

	return flagValuesMap
}
