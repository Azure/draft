package languages

import (
	"embed"
	"fmt"
	"io/fs"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

//go:generate cp -r ../../builders ./builders

var (
	//go:embed builders
	builders	embed.FS
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

	dir := "builders/" + val.Name()

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
	l := &Languages{
		langs: make(map[string]fs.DirEntry),
	}
	files, err := builders.ReadDir("builders")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.IsDir() {
			l.langs[f.Name()] = f
		}
	}

	return l
}
