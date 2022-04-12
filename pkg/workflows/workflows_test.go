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
		AksClusterName:    "test",
		ContainerName:     "test",
		ResourceGroupName: "test",

		chartsOverridePath: "testOverride",
		kustomizePath:      "testKustomize",
	}

	ghw := &GitHubWorkflow{}
	replaceWorkflowVars("", config, ghw)
	assert.NotNil(t, ghw.Env, "check that replace will update a ghw's environment")

	workflow, ok := deployNameToWorkflow["manifests"]
	assert.True(t, ok)

	ghw = getWorkflowFile(workflow)
	origLen := len(ghw.Jobs["build"].Steps)
	replaceWorkflowVars("manifests", config, ghw)
	assert.Equal(t, origLen-1, len(ghw.Jobs["build"].Steps), "check step is deleted")

	workflow, ok = deployNameToWorkflow["helm"]
	assert.True(t, ok)

	ghw = getWorkflowFile(workflow)
	replaceWorkflowVars("helm", config, ghw)
	assert.Equal(t, "testOverride", ghw.Env["CHART_OVERRIDE_PATH"], "check helm envs are replaced")

	workflow, ok = deployNameToWorkflow["kustomize"]
	assert.True(t, ok)

	ghw = getWorkflowFile(workflow)
	replaceWorkflowVars("kustomize", config, ghw)
	assert.Equal(t, "testKustomize", ghw.Env["KUSTOMIZE_PATH"], "check kustomize envs are replaces")
}

func TestUpdateProductionDeployments(t *testing.T) {
	config := &WorkflowConfig{
		AcrName:           "test",
		ContainerName:     "test",
		ResourceGroupName: "test",
	}
	assert.Nil(t, updateProductionDeployments("", ".", config))
}
