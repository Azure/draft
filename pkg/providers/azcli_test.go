package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAzCliInstalled(t *testing.T) {
	az := &AzClient{CommandRunner: &FakeCommandRunner{Output: `{
		"azure-cli": "2.65.0",
		"azure-cli-core": "2.65.0",
		"azure-cli-telemetry": "1.1.0",
		"extensions": {}
	}`}}
	assert.NotPanics(t, func() { az.ValidateAzCliInstalled() })

}
func TestCheckAzCliInstalledError(t *testing.T) {
	az := &AzClient{CommandRunner: &FakeCommandRunner{Output: "az", ErrStr: "error"}}
	err := az.ValidateAzCliInstalled()
	assert.NotNil(t, err)
}
