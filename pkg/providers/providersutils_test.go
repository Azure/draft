package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggedInToAz(t *testing.T) {
	isLoggedIn, err := IsLoggedInToAz()
	// In this case we should return false and receive an error.
	if err == nil {
		t.Error(err)
	}
	assert.False(t, isLoggedIn, "Azure CLI is returning logged in even when logged out")
}

func TestLoggedInToGh(t *testing.T) {
	assert.False(t, IsLoggedInToGh(), "GitHub CLI is returning logged in even when logged out")
}

func TestCheckAzCliInstalled(t *testing.T) {
	err := CheckAzCliInstalled()
	if err != nil {
		t.Error(err)
	}
}

func TestHasGhCli(t *testing.T) {
	assert.True(t, HasGhCli(), "GitHub CLI is not installed")
}
