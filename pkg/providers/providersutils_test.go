package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggedInToAz(t *testing.T) {
	isLoggedIn, err := IsLoggedInToAz()
	if err != nil {
		t.Error(err)
	}
	assert.False(t, isLoggedIn, "Azure CLI is returning logged in even when logged out")
}

func TestLoggedInToGh(t *testing.T) {
	assert.False(t, IsLoggedInToGh(), "Github is returning logged in even when logged out")
}

func TestCheckAzCliInstalled(t *testing.T) {
	err := CheckAzCliInstalled()
	if err != nil {
		t.Error(err)
	}
}

func TestHasGhCli(t *testing.T) {
	assert.True(t, HasGhCli(), "Github CLI is not installed")
}