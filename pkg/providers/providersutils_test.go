package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggedInToAz(t *testing.T) {
	assert.False(t, IsLoggedInToAz(), "AZ is returning logged in even when logged out")
}

func TestLoggedInToGh(t *testing.T) {
	assert.False(t, IsLoggedInToGh(), "Github is returning logged in even when logged out")
}

func TestCheckAzCliInstalled(t *testing.T) {
	var err error 
	CheckAzCliInstalled()

	assert.Nil(t, err)
}

func TestHasGhCli(t *testing.T) {
	assert.True(t, HasGhCli(), "Github CLI is not installed")
}