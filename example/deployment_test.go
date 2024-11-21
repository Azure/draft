package example

import (
	"testing"
)

func TestWriteDeploymentFilesExample(t *testing.T) {
	err := WriteDeploymentFilesExample()
	if err != nil {
		t.Errorf("WriteDockerfileExample failed: %e", err)
		t.Fail()
	}
}
