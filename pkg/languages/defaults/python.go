package defaults

import (
	"fmt"
	"strings"

	"github.com/Azure/draft/pkg/reporeader"
)

type PythonExtractor struct {
}

// ReadDefaults reads the default values for the language from the repo files
func (p PythonExtractor) ReadDefaults(r reporeader.RepoReader) (map[string]string, error) {
	extractedValues := make(map[string]string)
	files, err := r.FindFiles(".", []string{"*.py"}, 0)
	if err != nil {
		return nil, fmt.Errorf("error finding python files: %v", err)
	}
	for index, file := range files {
		fileContent, err := r.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf(("error reading python files"))
		}
		fileContentInString := string(fileContent)
		if strings.Contains(fileContentInString, `if __name__ == '__main__'`) || file == "main.py" || file == "app.py" {
			extractedValues["ENTRYPOINT"] = files[index]
			break
		}
	}

	if _, ok := extractedValues["ENTRYPOINT"]; !ok {
		if len(files) > 0 {
			extractedValues["ENTRYPOINT"] = files[0]
		}
	}

	return extractedValues, nil
}

func (p PythonExtractor) MatchesLanguage(lowerlang string) bool {
	return lowerlang == "python"
}

func (p PythonExtractor) GetName() string { return "python" }

var _ reporeader.VariableExtractor = &PythonExtractor{}
