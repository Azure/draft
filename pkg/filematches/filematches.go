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
        } else if matched {
            matches = append(matches, path)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return matches, nil
}

func isValidYamlFile(filePath string) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	results, err := kubeval.Validate(fileContents, filePath)
}