package workflows

import "github.com/Azure/draft/pkg/prompts"

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

type ServiceManifest interface {
	SetAnnotations(map[string]string)
	SetServiceType(string)
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

func (hpy *HelmProductionYaml) SetAnnotations(annotations map[string]string) {
	hpy.Service.Annotations = annotations
}

func (hpy *HelmProductionYaml) SetServiceType(serviceType string) {
	hpy.Service.ServiceType = serviceType
}

type DeploymentYaml struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   metadata `yaml:"metadata"`
	Spec       spec     `yaml:"spec"`
}

type metadata struct {
	Name        string            `yaml:"name"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

type spec struct {
	Template template               `yaml:"template"`
	Replicas string                 `yaml:"replicas"`
	Selector map[string]interface{} `yaml:"selector"`
}
type template struct {
	Spec     containers             `yaml:"spec"`
	Metadata map[string]interface{} `yaml:"metadata"`
}

type containers struct {
	Containers []container `yaml:"containers"`
}

type container struct {
	Name  string                   `yaml:"name"`
	Image string                   `yaml:"image"`
	Ports []map[string]interface{} `yaml:"ports"`
}

type ServiceYaml struct {
	ApiVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   metadata    `yaml:"metadata"`
	Spec       serviceSpec `yaml:"spec"`
}

type serviceSpec struct {
	ServiceType string                   `yaml:"type"`
	Selector    map[string]interface{}   `yaml:"selector,omitempty"`
	Ports       []map[string]interface{} `yaml:"ports,omitempty"`
}

func (sy *ServiceYaml) SetAnnotations(annotations map[string]string) {
	sy.Metadata.Annotations = annotations
}

func (sy *ServiceYaml) SetServiceType(serviceType string) {
	sy.Spec.ServiceType = serviceType
}
