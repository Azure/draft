package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	config := &WorkflowConfig{
		AcrName:           "Test",
		ContainerName:     "Test",
		ResourceGroupName: "test",
		AksClusterName:    "test",
		BranchName:        "test",
		BuildContextPath:  "./test",
	}
	config.ValidateAndFillConfig()
	assert.NotEmpty(t, config.KustomizePath)
	assert.NotEmpty(t, config.ManifestsPath)
	assert.NotEmpty(t, config.ChartsPath)
	assert.NotEmpty(t, config.ChartsOverridePath)
}
