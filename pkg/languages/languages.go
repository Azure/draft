package languages

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/imiller31/draftv2/pkg/embedutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/fs"
	"os"
	"strings"
)

//go:generate cp -r ../../builders ./builders

var (
	//go:embed builders
	builders	embed.FS
	parentDirName = "builders"
)

type Languages struct {
	langs	map[string]fs.DirEntry
	configs map[string]*Config
}

func (l *Languages) ContainsLanguage(lang string) bool {
	_, ok := l.langs[lang]
	return ok
}

func (l *Languages) CreateDockerfileForLanguage(lang string, customOverrides map[string]string) error {
	val, ok := l.langs[lang]
	if !ok {
		return fmt.Errorf("language %s is not supported", lang)
	}

	config := l.GetConfig(lang)

	dir := parentDirName + "/" + val.Name()

	files, err := fs.ReadDir(builders, dir)
	if err != nil {
		return err
	}

	log.Infof("got %d file matches", len(files))

	for _, f := range files {

		if f.Name() == "draft.yaml" {
			continue
		}

		filePath := dir + "/" + f.Name()

		log.Infof("fileName: %s", filePath)

		file, err := fs.ReadFile(builders, filePath)
		if err != nil {
			return err
		}

		fileString := string(file)

		for oldString, newString := range customOverrides {
			log.Debugf("replacing %s with %s", oldString, newString)
			fileString = strings.ReplaceAll(fileString, "{{" + oldString + "}}", newString)
		}

		fileName := f.Name()

		for _, fileOverride := range config.NameOverrides {
			fullPath := dir + "/" + fileOverride.Path
			log.Debugf("fullPath: %s, filePath: %s", fullPath, filePath)
			if fullPath == filePath {
				fileName = fileOverride.Prefix + fileName
				break
			}
		}

		log.Debugf("writing file: %s", fileName)

		if err = os.WriteFile(fileName, []byte(fileString), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (l *Languages) loadConfig(lang string) (*Config, error){
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
	viper.ReadConfig(bytes.NewBuffer(configBytes))

	var config Config

	if err = viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (l *Languages) GetConfig(lang string) *Config {
	val, ok := l.configs[lang]
	if !ok {
		return nil
	}
	return val
}

func (l *Languages) PopulateConfigs() {
	for lang, _ := range l.langs {
		config, err := l.loadConfig(lang)
		if err != nil {
			log.Debugf("no config found for language %s", lang)
			config = &Config{Language: lang}
		}
		l.configs[lang] = config
	}
}

func CreateLanguages() *Languages {
	langMap, err := embedutils.EmbedFStoMap(builders, parentDirName)
	if err != nil {
		log.Fatal(err)
	}


	l := &Languages{
		langs: langMap,
		configs: make(map[string]*Config),
	}
	l.PopulateConfigs()

	return l
}
