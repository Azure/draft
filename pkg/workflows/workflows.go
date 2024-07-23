package workflows

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/deployments"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

type Workflows struct {
	workflows         map[string]fs.DirEntry
	configs           map[string]*config.DraftConfig
	Dest              string
	workflowTemplates fs.FS
}

const (
	parentDirName         = "workflows"
	configFileName        = "/draft.yaml"
	emptyDefaultFlagValue = ""
)

func UpdateProductionDeployments(deployType, dest string, draftConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	deployment := deployments.CreateDeploymentsFromEmbedFS(template.Deployments, dest)
	var deployConfig *config.DraftConfig
	deployConfig, err := deployment.GetConfig(deployType)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	acr, err := draftConfig.GetVariable("AZURECONTAINERREGISTRY")
	if err != nil {
		return fmt.Errorf("get variable: %w", err)
	}

	containerName, err := draftConfig.GetVariable("CONTAINERNAME")
	if err != nil {
		return fmt.Errorf("get variable: %w", err)
	}

	productionImage := fmt.Sprintf("%s.azurecr.io/%s", acr.Value, containerName.Value)

	namespace, err := draftConfig.GetVariable("NAMESPACE")
	if err != nil {
		return fmt.Errorf("failed to get variable: %w", err)
	}

	deployConfig.SetVariable("IMAGENAME", productionImage)
	deployConfig.SetVariable("NAMESPACE", namespace.Value)

	if err = prompts.RunPromptsFromConfigWithSkips(deployConfig); err != nil {
		return fmt.Errorf("failed to run prompts from config with skips: %w", err)
	}

	if err = deployment.CopyDeploymentFiles(deployType, deployConfig, &writers.LocalFSWriter{}); err != nil {
		return fmt.Errorf("failed to copy deployment files: %w", err)
	}

	return nil
}

func (w *Workflows) loadConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := w.workflows[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}

	configPath := path.Join(parentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(w.workflowTemplates, configPath)
	if err != nil {
		return nil, err
	}

	var draftConfig config.DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
}

func (w *Workflows) GetConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := w.configs[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}
	return val, nil
}

func CreateWorkflowsFromEmbedFS(workflowTemplates embed.FS, dest string) *Workflows {
	deployMap, err := embedutils.EmbedFStoMap(workflowTemplates, parentDirName)
	if err != nil {
		log.Fatal(err)
	}

	w := &Workflows{
		workflows:         deployMap,
		Dest:              dest,
		configs:           make(map[string]*config.DraftConfig),
		workflowTemplates: workflowTemplates,
	}
	w.populateConfigs()

	return w
}

func (w *Workflows) populateConfigs() {
	for deployType := range w.workflows {
		draftConfig, err := w.loadConfig(deployType)
		if err != nil {
			log.Debugf("no draftConfig found for workflow of deploy type %s", deployType)
			draftConfig = &config.DraftConfig{}
		}
		w.configs[deployType] = draftConfig
	}
}

func (w *Workflows) CreateWorkflowFiles(deployType string, draftConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	val, ok := w.workflows[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}

	srcDir := path.Join(parentDirName, val.Name())
	log.Debugf("source directory for workflow template: %s", srcDir)

	if err := draftConfig.ApplyDefaultVariables(); err != nil {
		return fmt.Errorf("create workflow files: %w", err)
	}

	if err := osutil.CopyDir(w.workflowTemplates, srcDir, w.Dest, draftConfig, templateWriter); err != nil {
		return err
	}

	return nil
}
