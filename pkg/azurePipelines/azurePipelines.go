package azurePipelines

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	pipelineParentDirName       = "azurePipelines"
	aksPipelineTemplateFileName = "azure-kubernetes-service.yaml"
	configFileName              = "draft.yaml"
	pipelineNameVar             = "PIPELINENAME"
)

type AzurePipelines struct {
	pipelines         map[string]fs.DirEntry
	configs           map[string]*config.DraftConfig
	dest              string
	pipelineTemplates embed.FS
}

func CreatePipelinesFromEmbedFS(pipelineTemplates embed.FS, dest string) (*AzurePipelines, error) {
	pipelineMap, err := embedutils.EmbedFStoMap(pipelineTemplates, "azurePipelines")
	if err != nil {
		return nil, fmt.Errorf("error creating map from embedded FS: %w", err)
	}

	p := &AzurePipelines{
		pipelines:         pipelineMap,
		dest:              dest,
		configs:           make(map[string]*config.DraftConfig),
		pipelineTemplates: pipelineTemplates,
	}
	p.populateConfigs()

	return p, nil

}

func (p *AzurePipelines) populateConfigs() {
	for _, val := range p.pipelines {
		draftConfig, err := p.loadConfig(val.Name())
		if err != nil {
			log.Debugf("error loading draftConfig for pipeline of deploy type %s: %v", val.Name(), err)
			draftConfig = &config.DraftConfig{}
		}
		p.configs[val.Name()] = draftConfig
	}
}

func (p *AzurePipelines) GetConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := p.configs[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}
	return val, nil
}

func (p *AzurePipelines) loadConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := p.pipelines[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}

	configPath := path.Join(pipelineParentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(p.pipelineTemplates, configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var draftConfig config.DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %w", err)
	}

	return &draftConfig, nil
}

func (p *AzurePipelines) overrideFilename(draftConfig *config.DraftConfig, srcDir string) error {
	if draftConfig.FileNameOverrideMap == nil {
		draftConfig.FileNameOverrideMap = make(map[string]string)
	}
	pipelineVar, err := draftConfig.GetVariable(pipelineNameVar)
	if err != nil {
		return fmt.Errorf("error getting pipeline name variable: %w", err)
	}

	if err = fs.WalkDir(p.pipelineTemplates, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == aksPipelineTemplateFileName {
			draftConfig.FileNameOverrideMap[d.Name()] = pipelineVar.Value + ".yaml"
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error walking through source directory: %w", err)
	}

	return nil
}

func (p *AzurePipelines) CreatePipelineFiles(deployType string, draftConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	val, ok := p.pipelines[deployType]
	if !ok {
		return fmt.Errorf("deploy type %s currently unsupported for azure pipeline", deployType)
	}
	srcDir := path.Join(pipelineParentDirName, val.Name())
	log.Debugf("source directory of pipeline template: %s", srcDir)

	if err := p.overrideFilename(draftConfig, srcDir); err != nil {
		return fmt.Errorf("error overriding filename: %w", err)
	}

	if err := draftConfig.ApplyDefaultVariables(); err != nil {
		return fmt.Errorf("error applying default variables: %w", err)
	}

	if err := osutil.CopyDir(p.pipelineTemplates, srcDir, p.dest, draftConfig, templateWriter); err != nil {
		return fmt.Errorf("error copying pipeline files: %w", err)
	}

	return nil
}
