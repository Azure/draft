package reporeader

import (
	"path/filepath"
	"sort"
	"strings"
)

type RepoReader interface {
	Exists(path string) bool
	ReadFile(path string) ([]byte, error)
	// FindFiles returns a list of files that match the given patterns searching up to
	// maxDepth nested sub-directories. maxDepth of 0 limits files to the root dir.
	FindFiles(path string, patterns []string, maxDepth int) ([]string, error)
	GetRepoName() (string, error)
}

// VariableExtractor is an interface that can be implemented for extracting variables from a repo's files
type VariableExtractor interface {
	ReadDefaults(r RepoReader) (map[string]string, error)
	MatchesLanguage(lowerlang string) bool
	GetName() string
}

// FakeRepoReader is a RepoReader that can be used for testing, and takes a list of relative file paths with their contents
type FakeRepoReader struct {
	Files map[string][]byte
}

// GetRepoName returns the name of the repo
func (FakeRepoReader) GetRepoName() (string, error) {
	return "test-repo", nil
}

var _ RepoReader = FakeRepoReader{
	Files: map[string][]byte{},
}

func (r FakeRepoReader) Exists(path string) bool {
	if r.Files != nil {
		_, ok := r.Files[path]
		return ok
	}
	return false
}

func (r FakeRepoReader) ReadFile(path string) ([]byte, error) {
	if r.Files != nil {
		return r.Files[path], nil
	}
	return nil, nil
}

func (r FakeRepoReader) FindFiles(path string, patterns []string, maxDepth int) ([]string, error) {
	var files []string
	if r.Files == nil {
		return files, nil
	}

	// sort files because map iteration order is undefined. lets us control test behavior
	sortedFiles := make([]string, 0, len(r.Files))
	for k := range r.Files {
		sortedFiles = append(sortedFiles, k)
	}
	sort.Strings(sortedFiles)

	for _, file := range sortedFiles {
		for _, pattern := range patterns {
			if matched, err := filepath.Match(pattern, filepath.Base(file)); err != nil {
				return nil, err
			} else if matched {
				splitPath := strings.Split(file, string(filepath.Separator))
				fileDepth := len(splitPath) - 1
				if fileDepth <= maxDepth {
					files = append(files, file)
				}
			}
		}
	}
	return files, nil
}
