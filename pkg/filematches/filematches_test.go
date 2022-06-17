package filematches

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateYamlFromTemplate(dir string, valid bool) (*os.File, error) {
	file, err := ioutil.TempFile(dir, "*.yaml")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var source *os.File
	if valid {
		source, err = os.Open("./templates/valid_template.yaml")
	} else {
		source, err = os.Open("./templates/invalid_template.yaml")
	}
	if err != nil {
		return nil, err
	}
	defer source.Close()

	_, err = io.Copy(file, source)
	if err != nil {
		return nil, err
	}
	return file, nil
} 

func TestCreateK8sFileMatchesValidFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "filematch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file, err := generateYamlFromTemplate(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	file_name := file.Name()

	fileMatches, err := createK8sFileMatches(dir)
	if err != nil {
		t.Error(err)
	}
	assert.True(t, fileMatches.hasDeploymentFiles(), "should have valid deployment files")

	os.Remove(file_name)
	fileMatches, _ = createK8sFileMatches(dir)
	assert.False(t, fileMatches.hasDeploymentFiles(), "should not have valid deployment files")
}

func TestCreateK8sFileMatchesInvalidFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "filematch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	
	file, err := generateYamlFromTemplate(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	file_name := file.Name()

	fileMatches, _ := createK8sFileMatches(dir)
	assert.False(t, fileMatches.hasDeploymentFiles(), "should not have valid deployment files")

	os.Remove(file_name)
}

func TestCreateK8sFileMatchesNestedValidFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "filematch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	nestedDir, err := ioutil.TempDir(dir, "nested")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(nestedDir)
	
	file, err := generateYamlFromTemplate(nestedDir, true)
	if err != nil {
		t.Fatal(err)
	}
	file_name := file.Name()

	fileMatches, _ := createK8sFileMatches(dir)
	assert.True(t, fileMatches.hasDeploymentFiles(), "should have valid deployment files")

	os.Remove(file_name)
	fileMatches, _ = createK8sFileMatches(dir)
	assert.False(t, fileMatches.hasDeploymentFiles(), "should not have valid deployment files")
}

func TestCreateK8sFileMatchesNestedInvalidFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "filematch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	nestedDir, err := ioutil.TempDir(dir, "nested")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(nestedDir)

	file, err := generateYamlFromTemplate(nestedDir, false)
	if err != nil {
		t.Fatal(err)
	}
	file_name := file.Name()

	fileMatches, _ := createK8sFileMatches(dir)
	assert.False(t, fileMatches.hasDeploymentFiles(), "should not have valid deployment files")

	os.Remove(file_name)
}

func touchDockerfile(name string) error {
    file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    return file.Close()
}

func TestSearchDirectoryWithDockerfile(t *testing.T) {
	dir, err := ioutil.TempDir("", "filematch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	dockerfilePath := dir + "/Dockerfile"
	err = touchDockerfile(dockerfilePath)
	if err != nil {
		t.Fatal(err)
	}

	hasDockerFile, _, err := SearchDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, hasDockerFile, "should have Dockerfile")

	os.Remove(dockerfilePath)
	hasDockerFile, _, err = SearchDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, hasDockerFile, "should not have Dockerfile")
}