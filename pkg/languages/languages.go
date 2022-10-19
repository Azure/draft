package languages

import (
	"embed"
	"fmt"
	"io/fs"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter"
)

var (
	parentDirName = "dockerfiles"
)

type Languages struct {
	langs               map[string]fs.DirEntry
	configs             map[string]*config.DraftConfig
	dest                string
	dockerfileTemplates fs.FS
}

func (l *Languages) ContainsLanguage(lang string) bool {
	_, ok := l.langs[lang]
	return ok
}

func (l *Languages) CreateDockerfileForLanguage(lang string, customInputs map[string]string, templateWriter templatewriter.TemplateWriter) error {
	val, ok := l.langs[lang]
	if !ok {
		return fmt.Errorf("language %s is not supported", lang)
	}

	srcDir := parentDirName + "/" + val.Name()

	draftConfig, ok := l.configs[lang]
	if !ok {
		draftConfig = nil
	}

	if err := osutil.CopyDir(l.dockerfileTemplates, srcDir, l.dest, draftConfig, customInputs, templateWriter); err != nil {
		return err
	}

	return nil
}

func (l *Languages) loadConfig(lang string) (*config.DraftConfig, error) {
	val, ok := l.langs[lang]
	if !ok {
		return nil, fmt.Errorf("language %s unsupported", lang)
	}

	configPath := parentDirName + "/" + val.Name() + "/draft.yaml"
	configBytes, err := fs.ReadFile(l.dockerfileTemplates, configPath)
	if err != nil {
		return nil, err
	}

	var draftConfig config.DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
}

func (l *Languages) GetConfig(lang string) *config.DraftConfig {
	val, ok := l.configs[lang]
	if !ok {
		return nil
	}
	return val
}

func (l *Languages) PopulateConfigs() {
	for lang := range l.langs {
		draftConfig, err := l.loadConfig(lang)
		if err != nil {
			log.Debugf("no draftConfig found for language %s", lang)
			draftConfig = &config.DraftConfig{}
		}
		l.configs[lang] = draftConfig
	}
}

func CreateLanguagesFromEmbedFS(dockerfileTemplates embed.FS, dest string) *Languages {
	langMap, err := embedutils.EmbedFStoMap(dockerfileTemplates, parentDirName)
	if err != nil {
		log.Fatal(err)
	}

	l := &Languages{
		langs:               langMap,
		dest:                dest,
		configs:             make(map[string]*config.DraftConfig),
		dockerfileTemplates: dockerfileTemplates,
	}
	l.PopulateConfigs()

	return l
}
