package reporeader

import (
	"path/filepath"
	"strings"
)

type RepoReader interface {
	Exists(path string) bool
	ReadFile(path string) ([]byte, error)
	// FindFiles returns a list of files that match the given patterns searching up to
	// maxDepth nested sub-directories. maxDepth of 0 limits files to the root dir.
	FindFiles(path string, patterns []string, maxDepth int) ([]string, error)
}

// VariableExtractor is an interface that can be implemented for extracting variables from a repo's files
type VariableExtractor interface {
	ReadDefaults(r RepoReader, dest string) (map[string]string, error)
	MatchesLanguage(lowerlang string) bool
	GetName() string
}

// TestRepoReader is a RepoReader that can be used for testing, and takes a list of relative file paths with their contents
type TestRepoReader struct {
	Files map[string][]byte
}

func (r TestRepoReader) Exists(path string) bool {
	if r.Files != nil {
		_, ok := r.Files[path]
		return ok
	}
	return false
}

func (r TestRepoReader) ReadFile(path string) ([]byte, error) {
	if r.Files != nil {
		return r.Files[path], nil
	}
	return nil, nil
}

func (r TestRepoReader) FindFiles(path string, patterns []string, maxDepth int) ([]string, error) {
	var files []string
	if r.Files == nil {
		return files, nil
	}
	for k := range r.Files {
		for _, pattern := range patterns {
			if matched, err := filepath.Match(pattern, filepath.Base(k)); err != nil {
				return nil, err
			} else if matched {
				splitPath := strings.Split(k, string(filepath.Separator))
				fileDepth := len(splitPath) - 1
				if fileDepth <= maxDepth {
					files = append(files, k)
				}
			}
		}
	}
	return files, nil
}
