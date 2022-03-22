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

func TestCreateFileMatchesValidFile(t *testing.T) {
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

	fileMatches := CreateFileMatches(dir)
	assert.True(t, fileMatches.HasDeploymentFiles(), "should have valid deployment files")

	os.Remove(file_name)
	fileMatches = CreateFileMatches(dir)
	assert.False(t, fileMatches.HasDeploymentFiles(), "should not have valid deployment files")
}

func TestCreateFileMatchesInvalidFile(t *testing.T) {
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

	fileMatches := CreateFileMatches(dir)
	assert.False(t, fileMatches.HasDeploymentFiles(), "should not have valid deployment files")

	os.Remove(file_name)
}

func TestCreateFileMatchesNestedValidFile(t *testing.T) {
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

	fileMatches := CreateFileMatches(dir)
	assert.True(t, fileMatches.HasDeploymentFiles(), "should have valid deployment files")

	os.Remove(file_name)
	fileMatches = CreateFileMatches(dir)
	assert.False(t, fileMatches.HasDeploymentFiles(), "should not have valid deployment files")
}

func TestCreateFileMatchesNestedInvalidFile(t *testing.T) {
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

	fileMatches := CreateFileMatches(dir)
	assert.False(t, fileMatches.HasDeploymentFiles(), "should not have valid deployment files")

	os.Remove(file_name)
}
