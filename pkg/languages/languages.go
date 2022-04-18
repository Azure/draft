package languages

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"

	"github.com/Azure/draft/pkg/configs"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:generate cp -r ../../builders ./builders

var (
	//go:embed all:builders
	builders      embed.FS
	parentDirName = "builders"
)

type Languages struct {
	langs   map[string]fs.DirEntry
	configs map[string]*configs.DraftConfig
	dest    string
}

func (l *Languages) ContainsLanguage(lang string) bool {
	_, ok := l.langs[lang]
	return ok
}

func (l *Languages) CreateDockerfileForLanguage(lang string, customInputs map[string]string) error {
	val, ok := l.langs[lang]
	if !ok {
		return fmt.Errorf("language %s is not supported", lang)
	}

	srcDir := parentDirName + "/" + val.Name()

	config, ok := l.configs[lang]
	if !ok {
		config = nil
	}

	if err := osutil.CopyDir(builders, srcDir, l.dest, config, customInputs); err != nil {
		return err
	}

	return nil
}

func (l *Languages) loadConfig(lang string) (*configs.DraftConfig, error) {
	val, ok := l.langs[lang]
	if !ok {
		return nil, fmt.Errorf("language %s unsupported", lang)
	}

	configPath := parentDirName + "/" + val.Name() + "/draft.yaml"
	configBytes, err := fs.ReadFile(builders, configPath)
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

func (l *Languages) GetConfig(lang string) *configs.DraftConfig {
	val, ok := l.configs[lang]
	if !ok {
		return nil
	}
	return val
}

func (l *Languages) PopulateConfigs() {
	for lang := range l.langs {
		config, err := l.loadConfig(lang)
		if err != nil {
			log.Debugf("no config found for language %s", lang)
			config = &configs.DraftConfig{}
		}
		l.configs[lang] = config
	}
}

func CreateLanguages(dest string) *Languages {
	langMap, err := embedutils.EmbedFStoMap(builders, parentDirName)
	if err != nil {
		log.Fatal(err)
	}

	l := &Languages{
		langs:   langMap,
		dest:    dest,
		configs: make(map[string]*configs.DraftConfig),
	}
	l.PopulateConfigs()

	return l
}
