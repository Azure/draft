package defaults

import (
	"fmt"

	"github.com/Azure/draft/cmd"
	"github.com/Azure/draft/pkg/reporeader"
)

type PythonExtractor struct {
	cmd.CreateConfig
}

// ReadDefaults reads the default values for the language from the repo files
func (p PythonExtractor) ReadDefaults(r reporeader.RepoReader) (map[string]string, error) {
	extractedValues := make(map[string]string)
	files, err := r.FindFiles(".", []string{"*.py"}, 2)
	if err != nil {
		return nil, fmt.Errorf("error finding python files: %v", err)
	}
	if len(files) > 0 {
		extractedValues["ENTRYPOINT"] = files[0]
	}

	return extractedValues, nil
}

func (p PythonExtractor) MatchesLanguage(lowerlang string) bool {
	return lowerlang == "python"
}

func (p PythonExtractor) GetName() string { return "python" }

var _ reporeader.VariableExtractor = &PythonExtractor{}
