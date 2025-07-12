package example

import (
	"testing"
)

func TestWriteClusterResourcePlacementFilesExample(t *testing.T) {
	err := WriteClusterResourcePlacementFilesExample()
	if err != nil {
		t.Errorf("WriteClusterResourcePlacementFilesExample failed: %e", err)
		t.Fail()
	}
}