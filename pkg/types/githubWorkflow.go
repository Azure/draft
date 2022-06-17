package types

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
	RunsOn      string   `yaml:"runs-on"`
	Needs       []string `yaml:"needs,omitempty"`
	Steps       []map[string]interface{}
}
