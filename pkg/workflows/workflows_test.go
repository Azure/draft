package workflows

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorkflowEmbed(t *testing.T) {
	workflow := &workflowType{
		deployPath:         "/charts",
		workflowFileSuffix: "-helm",
	}

	assert.NotEmptyf(t, getWorkflowFile(workflow), "workflow should be fetched from the embeded file system")
}

func TestWorkflowReplace(t *testing.T) {
	config := &WorkflowConfig{
		AcrName:           "test",
		ContainerName:     "test",
		ResourceGroupName: "test",
	}

	assert.Equal(t, "testing",
		replaceWorkflowVars("your-azure-container-registrying", config))

	assert.Equal(t, "nochange",
		replaceWorkflowVars("nochange", config))
}
