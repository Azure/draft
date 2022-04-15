package providers

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggedInToAz(t *testing.T) {
	CheckAzCliInstalled()

	logOutCmd := exec.Command("az", "logout")
	out, err := logOutCmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
	}

	assert.False(t, IsLoggedInToAz(), "AZ is returning logged in even when logged out")
}
