package readers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/reporeader"
)

type LocalFSReader struct {
}

// GetRepoName returns the name of the current directory, which is an approximation of the repo name
func (*LocalFSReader) GetRepoName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get working directory: %v", err)
	}
	dirName := filepath.Base(wd)
	return dirName, nil
}

type LocalFileFinder struct {
	Patterns   []string
	FoundFiles []string
	MaxDepth   int
}

func (l *LocalFileFinder) walkFunc(path string, info os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	// Skip directories that are too deep
	if info.IsDir() && strings.Count(path, string(os.PathSeparator)) > l.MaxDepth {
		fmt.Println("skip", path)
		return fs.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	for _, pattern := range l.Patterns {
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			l.FoundFiles = append(l.FoundFiles, path)
		}
	}
	return nil
}

func (r *LocalFSReader) FindFiles(path string, patterns []string, maxDepth int) ([]string, error) {
	l := LocalFileFinder{
		Patterns: patterns,
		MaxDepth: maxDepth,
	}
	err := filepath.WalkDir(path, l.walkFunc)
	if err != nil {
		return nil, err
	}
	return l.FoundFiles, nil
}

var _ reporeader.RepoReader = &LocalFSReader{}

func (r *LocalFSReader) Exists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func (r *LocalFSReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
