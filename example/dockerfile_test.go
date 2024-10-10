package example

import (
	"testing"
)

func TestWriteDockerfileExample(t *testing.T) {
	err := WriteDockerfileExample()
	if err != nil {
		t.Errorf("WriteDockerfileExample failed: %e", err)
	}
}
