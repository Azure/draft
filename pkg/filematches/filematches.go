package filematches

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/instrumenta/kubeval/kubeval"
)

type FileMatches struct {
	dest            string
	patterns        []string
	deploymentFiles []string
}

func (f *FileMatches) findDeploymentFiles(dest string) error {
	err := filepath.Walk(dest, f.walkFunc)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileMatches) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatal(err)
		return err
	}
	if info.IsDir() {
		return nil
	}
	for _, pattern := range f.patterns {
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched && isValidK8sFile(path) {
			f.deploymentFiles = append(f.deploymentFiles, path)
		}
	}
	return nil
}

// TODO: maybe generalize this function in the future
func isValidK8sFile(filePath string) bool {
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	config := kubeval.NewDefaultConfig()
	results, err := kubeval.Validate(fileContents, config)
	if err != nil || hasErrors(results) {
		return false
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

func (f *FileMatches) hasDeploymentFiles() bool {
	return len(f.deploymentFiles) > 0
}

func createK8sFileMatches(dest string) *FileMatches {
	l := &FileMatches{
		dest:            dest,
		patterns:        []string{"*.yaml", "*.yml"},
		deploymentFiles: []string{},
	}
	err := l.findDeploymentFiles(dest)
	if err != nil {
		log.Fatal(err)
	}

	return l
}

func SearchDirectory(dest string) (bool, bool, error) {
	// check if Dockerfile exists
	var hasDockerFile bool
	dockerfilePath := dest + "/Dockerfile"
	_, err := os.Stat(dockerfilePath)
	if err == nil {
		hasDockerFile = true
	} else if errors.Is(err, os.ErrNotExist) {
		hasDockerFile = false
	} else {
		return false, false, err
	}

	// recursive directory search for valid yaml files
	fileMatches := createK8sFileMatches(dest)
	_, err = FindDraftDeploymentFiles(dest)
	hasDeploymentFiles := fileMatches.hasDeploymentFiles() || err == nil
	return hasDockerFile, hasDeploymentFiles, nil
}

func FindDraftDeploymentFiles(dest string) (deploymentType string, err error) {
	if _, err := os.Stat(dest + "/charts"); !os.IsNotExist(err) {
		return "helm", nil
	}
	if _, err := os.Stat(dest + "/overlays"); !os.IsNotExist(err) {
		return "kustomize", nil
	}
	if _, err := os.Stat(dest + "/manifests"); !os.IsNotExist(err) {
		return "manifests", nil
	}

	return "", errors.New("no supported deployment files found")
}
