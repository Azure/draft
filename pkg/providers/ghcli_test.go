package providers

import (
	"testing"
)

func TestHasGhCli(t *testing.T) {
	EnsureGhCliInstalled()
}
