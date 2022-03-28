package cmd

import (
	"testing"
)

func TestAzConfigExists (t *testing.T) {
	sample := &SetUpCmd{}
	_, err := sample.InitializeSetUpConfig()

	if err != nil {
		t.Error("Struct was not created correctly")
	}
}