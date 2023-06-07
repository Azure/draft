package reporeader

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
