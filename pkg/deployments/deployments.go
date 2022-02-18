package deployments

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"

	"github.com/imiller31/draftv2/pkg/configs"
	"github.com/imiller31/draftv2/pkg/embedutils"
	"github.com/imiller31/draftv2/pkg/osutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:generate cp -r ../../deployTypes ./deployTypes

var (
	//go:embed all:deployTypes
	deployTypes    embed.FS
	parentDirName  = "deployTypes"
	configFileName = "draft.yaml"
)

type Deployments struct {
	deploys map[string]fs.DirEntry
	configs map[string]*configs.DraftConfig
	dest    string
}

func (d *Deployments) CopyDeploymentFiles(deployType string, customInputs map[string]string) error {
	val, ok := d.deploys[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}

	srcDir := parentDirName + "/" + val.Name()

	config, ok := d.configs[deployType]
	if !ok {
		config = nil
	}

	if err := osutil.CopyDir(deployTypes, srcDir, d.dest, config, customInputs); err != nil {
		return err
	}

	return nil
}

func (d *Deployments) loadConfig(lang string) (*configs.DraftConfig, error) {
	val, ok := d.deploys[lang]
	if !ok {
		return nil, fmt.Errorf("language %s unsupported", lang)
	}

	configPath := fmt.Sprintf("%s/%s/%s", parentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(deployTypes, configPath)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile("yaml")
	if err = viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
		return nil, err
	}

	var config configs.DraftConfig

	if err = viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (d *Deployments) GetConfig(deployTyoe string) *configs.DraftConfig {
	val, ok := d.configs[deployTyoe]
	if !ok {
		return nil
	}
	return val
}

func (d *Deployments) PopulateConfigs() {
	for deployType := range d.deploys {
		config, err := d.loadConfig(deployType)
		if err != nil {
			log.Debugf("no config found for language %s", deployType)
			config = &configs.DraftConfig{}
		}
		d.configs[deployType] = config
	}
}

func CreateDeployments(dest string) *Deployments {
	deployMap, err := embedutils.EmbedFStoMap(deployTypes, "deployTypes")
	if err != nil {
		log.Fatal(err)
	}

	d := &Deployments{
		deploys: deployMap,
		dest:    dest,
		configs: make(map[string]*configs.DraftConfig),
	}
	d.PopulateConfigs()

	return d
}
