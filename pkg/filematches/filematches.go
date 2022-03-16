package filematches

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/instrumenta/kubeval/kubeval"
)

type FileMatches struct {
	dest   			string
	deploymentFiles []string
}

func findDeploymentFiles(dest string, pattern string) ([]string, error) {
    var matches []string
    err := filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil  
        }
        if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
            return err
        } else if matched && isValidYamlFile(path) {
            matches = append(matches, path)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return matches, nil
}

func isValidYamlFile(filePath string) bool {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
    config := kubeval.NewDefaultConfig()
	results, err := kubeval.Validate(fileContents, config)
    if err != nil || hasErrors(results) {
        log.Fatal(err)
    }
    return true
}

func hasErrors(res []kubeval.ValidationResult) bool {
	for _, r := range res {
		if len(r.Errors) > 0 {
			return true
		}
	}
	return false
}

func (f *FileMatches) HasDeploymentFiles() bool {
    return len(f.deploymentFiles) > 0
}

func CreateFileMatches(dest string) *FileMatches {
	deploymentFiles, err := findDeploymentFiles(dest, "*.yaml")
	if err != nil {
		log.Fatal(err)
	}

	l := &FileMatches{
		dest:            dest,
		deploymentFiles: deploymentFiles,
	}

	return l
}
