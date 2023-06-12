package reporeader

import "path/filepath"

type RepoReader interface {
	Exists(path string) bool
	ReadFile(path string) ([]byte, error)
	FindFiles(path string, patterns []string, maxDepth int) ([]string, error)
}

// VariableExtractor is an interface that can be implemented for extracting variables from a repo's files
type VariableExtractor interface {
	ReadDefaults(r RepoReader) (map[string]string, error)
	MatchesLanguage(lowerlang string) bool
	GetName() string
}

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
				fileDepth := len(filepath.SplitList(k))
				if fileDepth < maxDepth {
					files = append(files, k)
				}
			}
		}
	}
	return files, nil
}
