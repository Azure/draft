package deployments

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter"
)

var (
	parentDirName  = "deployments"
	configFileName = "draft.yaml"
)

type Deployments struct {
	deploys             map[string]fs.DirEntry
	configs             map[string]*config.DraftConfig
	dest                string
	deploymentTemplates fs.FS
}

func (d *Deployments) CopyDeploymentFiles(deployType string, customInputs map[string]string, templateWriter templatewriter.TemplateWriter) error {
	val, ok := d.deploys[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}

	srcDir := parentDirName + "/" + val.Name()

	config, ok := d.configs[deployType]
	if !ok {
		config = nil
	}

	if err := osutil.CopyDir(d.deploymentTemplates, srcDir, d.dest, config, customInputs, templateWriter); err != nil {
		return err
	}

	return nil
}

func (d *Deployments) loadConfig(lang string) (*config.DraftConfig, error) {
	val, ok := d.deploys[lang]
	if !ok {
		return nil, fmt.Errorf("language %s unsupported", lang)
	}

	configPath := fmt.Sprintf("%s/%s/%s", parentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(d.deploymentTemplates, configPath)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile("yaml")
	if err = viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
		return nil, err
	}

	var draftConfig config.DraftConfig

	if err = viper.Unmarshal(&draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
}

func (d *Deployments) GetConfig(deployType string) *config.DraftConfig {
	val, ok := d.configs[deployType]
	if !ok {
		return nil
	}
	return val
}

func (d *Deployments) PopulateConfigs() {
	for deployType := range d.deploys {
		draftConfig, err := d.loadConfig(deployType)
		if err != nil {
			log.Debugf("no draftConfig found for language %s", deployType)
			draftConfig = &config.DraftConfig{}
		}
		d.configs[deployType] = draftConfig
	}
}

func CreateDeploymentsFromEmbedFS(deploymentTemplates embed.FS, dest string) *Deployments {
	deployMap, err := embedutils.EmbedFStoMap(deploymentTemplates, "deployments")
	if err != nil {
		log.Fatal(err)
	}

	d := &Deployments{
		deploys:             deployMap,
		dest:                dest,
		configs:             make(map[string]*config.DraftConfig),
		deploymentTemplates: deploymentTemplates,
	}
	d.PopulateConfigs()

	return d
}
