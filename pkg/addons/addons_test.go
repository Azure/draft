package addons

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateAddon(t *testing.T) {
	userInputs := map[string]string{
		"test": "test",
	}
	err := GenerateAddon("azure", "webapp_routing", "baddest", userInputs)
	assert.NotNil(t, err)

	err = GenerateAddon("azure", "fakeaddon", "../../test/templates/helm", userInputs)
	assert.NotNil(t, err)

	err = GenerateAddon("fakeProvider", "fakeaddon", "../../test/templates/helm", userInputs)
	assert.NotNil(t, err)
}
