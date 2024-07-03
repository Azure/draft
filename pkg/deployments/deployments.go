package deployments

import (
	"embed"
	"fmt"
	"io/fs"
	"path"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

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

// DeployTypes returns a slice of the supported deployment types
func (d *Deployments) DeployTypes() []string {
	names := maps.Keys(d.deploys)
	return names
}

func (d *Deployments) CopyDeploymentFiles(deployType string, deployConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	val, ok := d.deploys[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}

	srcDir := path.Join(parentDirName, val.Name())

	if err := deployConfig.ApplyDefaultVariables(); err != nil {
		return fmt.Errorf("create deployment files for deployment type: %w", err)
	}

	if err := osutil.CopyDir(d.deploymentTemplates, srcDir, d.dest, deployConfig, templateWriter); err != nil {
		return err
	}

	return nil
}

func (d *Deployments) loadConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := d.deploys[deployType]
	if !ok {
		return nil, fmt.Errorf("deployment type %s unsupported", deployType)
	}

	configPath := path.Join(parentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(d.deploymentTemplates, configPath)
	if err != nil {
		return nil, err
	}

	var draftConfig config.DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
}

func (d *Deployments) GetConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := d.configs[deployType]
	if !ok {
		return nil, fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}
	return val, nil
}

func (d *Deployments) PopulateConfigs() {
	for deployType := range d.deploys {
		draftConfig, err := d.loadConfig(deployType)
		if err != nil {
			log.Debugf("no draftConfig found for deployment type %s", deployType)
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
