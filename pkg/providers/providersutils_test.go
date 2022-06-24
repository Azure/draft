package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggedInToAz(t *testing.T) {
	assert.NotNil(t, IsLoggedInToAz(), "Azure CLI is returning logged in even when logged out")
}

func TestLoggedInToGh(t *testing.T) {
	assert.NotNil(t, IsLoggedInToGh(), "GitHub CLI is returning logged in even when logged out")
}

func TestCheckAzCliInstalled(t *testing.T) {
	assert.Nil(t, CheckAzCliInstalled())
}

func TestHasGhCli(t *testing.T) {
	assert.True(t, HasGhCli(), "GitHub CLI is not installed")
}
