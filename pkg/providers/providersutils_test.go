package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAzCliInstalled(t *testing.T) {
	var err error
	CheckAzCliInstalled()

	assert.Nil(t, err)
}

func TestHasGhCli(t *testing.T) {
	assert.True(t, HasGhCli(), "Github CLI is not installed")
}
