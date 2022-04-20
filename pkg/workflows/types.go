package workflows

import (
	"github.com/Azure/draft/pkg/prompts"
)

//GitHubWorkflow is a rough struct to allow for yaml editing including deletion of Job steps
type GitHubWorkflow struct {
	Name string
	On   on `yaml:"on"`
	Env  map[string]string
	Jobs map[string]job
}

type on struct {
	Push             push
	WorkflowDispatch interface{} `yaml:"workflow_dispatch"`
}

type push struct {
	Branches []string
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
	BranchName         string
	manifestsPath      string
	chartsPath         string
	chartsOverridePath string
	kustomizePath      string
}

type HelmProductionYaml struct {
	ImageKey imageKey `yaml:"imageKey"`
	Service  service  `yaml:"service"`
}

type service struct {
	Annotations map[string]string `yaml:"annotations"`
	ServiceType string            `yaml:"type"`
	Port        string            `yaml:"port"`
}

type imageKey struct {
	Repository string `yaml:"repository"`
	PullPolicy string `yaml:"pullPolicy"`
	Tag        string `yaml:"tag"`
}

type DeploymentYaml struct {
	ApiVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       spec                   `yaml:"spec"`
}

type spec struct {
	Template template               `yaml:"template"`
	Replicas string                 `yaml:"replicas"`
	Selector map[string]interface{} `yaml:"selector"`
}
type template struct {
	Spec containers `yaml:"spec"`
}

type containers struct {
	Containers []container `yaml:"containers"`
}

type container struct {
	Name  string                   `yaml:"name"`
	Image string                   `yaml:"image"`
	Ports []map[string]interface{} `yaml:"ports"`
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

	if config.BranchName == "" {
		config.BranchName = prompts.GetInputFromPrompt("name of the repository branch to deploy from, usually main")
	}

	config.chartsPath = "./charts"
	config.chartsOverridePath = "./charts/production.yaml"
	config.manifestsPath = "./manifests"
	config.kustomizePath = "./overlays/production"
}
