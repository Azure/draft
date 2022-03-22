package filematches

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateFileMatches(t *testing.T) {
	valid_file, err := ioutil.TempFile("", "valid_file.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(valid_file)
	defer valid_file.Close()

	valid_source, err := io.Open("templates/valid_template.yaml")
	if err != nil {
			return 0, err
	}
	defer valid_source.Close()

	io.Copy(valid_source, valid_file)
	
	fileMatches := CreateFileMatches("")
	assert.Equal(fileMatches.HasDeploymentFiles(), true, "should have deployment files")
}