package providers

import (
	"testing"
)

func TestHasGhCli(t *testing.T) {
	cr := &FakeCommandRunner{
		Output: "gh version 1.0.0",
	}
	gh := &GhCliClient{
		CommandRunner: cr,
	}
	gh.EnsureGhCliInstalled()
}
