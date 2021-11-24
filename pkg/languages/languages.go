package languages

import (
	"embed"
	"fmt"
	"github.com/imiller31/draftv2/pkg/embedutils"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path"
)

//go:generate cp -r ../../builders ./builders

var (
	//go:embed builders
	builders	embed.FS
	parentDirName = "builders"
)

type Languages struct {
	langs	map[string]fs.DirEntry
}

func (l *Languages) ContainsLanguage(lang string) bool {
	_, ok := l.langs[lang]
	return ok
}

func (l *Languages) CreateDockerfileForLanguage(lang string) error {
	val, ok := l.langs[lang]
	if !ok {
		return fmt.Errorf("language %s is not supported", lang)
	}

	dir := parentDirName + "/" + val.Name()

	files, err := fs.ReadDir(builders, dir)
	if err != nil {
		return err
	}

	log.Infof("got %d file matches", len(files))

	for _, f := range files {

		filePath := dir + "/" + f.Name()

		log.Infof("fileName: %s", filePath)

		file, err := fs.ReadFile(builders, filePath)
		if err != nil {
			return err
		}

		if err = os.WriteFile(path.Base(f.Name()), file, 0644); err != nil {
			return err
		}
	}

	return nil
}

func CreateLanguages() *Languages {
	langMap, err := embedutils.EmbedFStoMap(builders, parentDirName)
	if err != nil {
		log.Fatal(err)
	}
	return &Languages{langs: langMap}
}
