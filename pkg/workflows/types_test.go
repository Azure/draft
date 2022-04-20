package workflows

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	config := &WorkflowConfig{
		AcrName:           "test",
		ContainerName:     "test",
		ResourceGroupName: "test",
		AksClusterName:    "test",
		BranchName:        "test",
	}
	config.ValidateAndFillConfig()
	assert.NotEmpty(t, config.kustomizePath)
	assert.NotEmpty(t, config.manifestsPath)
	assert.NotEmpty(t, config.chartsPath)
	assert.NotEmpty(t, config.chartsOverridePath)
}
