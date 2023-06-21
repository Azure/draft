package defaults

import (
	"fmt"
	"regexp"

	"github.com/Azure/draft/pkg/reporeader"
)

type PythonExtractor struct {
}

// ReadDefaults reads the default values for the language from the repo files
func (p PythonExtractor) ReadDefaults(r reporeader.RepoReader, dest string) (map[string]string, error) {
	extractedValues := make(map[string]string)
	files, err := r.FindFiles(dest, []string{"*.py"}, 0)
	if err != nil {
		return nil, fmt.Errorf("error finding python files: %v", err)
	}

	entryPointPattern := `if\s*__name__\s*==\s*["']__main__["']`
	compiledPattern, err := regexp.Compile(entryPointPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex pattern: %v", err)
	}
	for index, fileName := range files {
		fileContent, err := r.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf(("error reading python files"))
		}
		fileContentInString := string(fileContent)

		if compiledPattern.MatchString(fileContentInString) || fileName == "main.py" || fileName == "app.py" {
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
