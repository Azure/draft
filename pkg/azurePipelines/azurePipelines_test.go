package azurePipelines

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

func TestCreatePipelines(t *testing.T) {
	var pipelineFilePath string
	templateWriter := &writers.LocalFSWriter{}

	tests := []struct {
		name                    string
		deployType              string
		shouldError             bool
		setConfig               func(dc *config.DraftConfig)
		defaultPipelineFileName string
		cleanUp                 func(tempDir string)
	}{
		{
			name:                    "kustomize_default_path",
			deployType:              "kustomize",
			shouldError:             false,
			defaultPipelineFileName: "azure-kubernetes-service-kustomize.yaml",
		},
		{
			name:        "kustomize_given_path",
			deployType:  "kustomize",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("KUSTOMIZEPATH", "test/kustomize/overlays/production")
			},
			defaultPipelineFileName: "azure-kubernetes-service-kustomize.yaml",
		},
		{
			name:        "manifests_default_path",
			deployType:  "manifests",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("PIPELINENAME", "some-other-name")
			},
			defaultPipelineFileName: "azure-kubernetes-service.yaml",
		},
		{
			name:        "manifests_custom_path",
			deployType:  "manifests",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("MANIFESTPATH", "test/manifests")
			},
			defaultPipelineFileName: "azure-kubernetes-service.yaml",
		},
		{
			name:                    "invalid",
			deployType:              "invalid",
			shouldError:             true,
			defaultPipelineFileName: "azure-kubernetes-service.yaml",
		},
		{
			name:        "missing_config",
			deployType:  "kustomize",
			shouldError: true,
			setConfig: func(dc *config.DraftConfig) {
				// removing the last variable from draftConfig
				dc.Variables = dc.Variables[:len(dc.Variables)-1]
			},
			defaultPipelineFileName: "azure-kubernetes-service-kustomize.yaml",
		},
	}

	for _, tt := range tests {
		draftConfig := newDraftConfig()

		tempDir, err := os.MkdirTemp(".", "testTempDir")
		assert.Nil(t, err)

		if tt.setConfig != nil {
			tt.setConfig(draftConfig)
		}

		pipelines, err := CreatePipelinesFromEmbedFS(template.AzurePipelines, tempDir)
		assert.Nil(t, err)

		err = pipelines.CreatePipelineFiles(tt.deployType, draftConfig, templateWriter)

		pipelineFilePath = fmt.Sprintf("%s/.pipelines/%s", tempDir, tt.defaultPipelineFileName)
		if val, ok := draftConfig.FileNameOverrideMap[tt.defaultPipelineFileName]; ok {
			pipelineFilePath = fmt.Sprintf("%s/.pipelines/%s", tempDir, val)
		}

		if tt.shouldError {
			assert.NotNil(t, err)
			_, err = os.Stat(pipelineFilePath)
			assert.Equal(t, os.IsNotExist(err), true)
		} else {
			assert.Nil(t, err)
			_, err = os.Stat(pipelineFilePath)
			assert.Nil(t, err)
		}

		err = os.RemoveAll(tempDir)
		assert.Nil(t, err)
	}
}

func newDraftConfig() *config.DraftConfig {
	return &config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name:  "PIPELINENAME",
				Value: "testPipeline",
			},
			{
				Name: "BRANCHNAME",
				Default: config.BuilderVarDefault{
					Value: "main",
				},
			},
			{
				Name:  "ARMSERVICECONNECTION",
				Value: "testServiceConnection",
			},
			{
				Name:  "AZURECONTAINERREGISTRY",
				Value: "testACR",
			},
			{
				Name:  "CONTAINERNAME",
				Value: "testContainer",
			},
			{
				Name:  "CLUSTERRESOURCEGROUP",
				Value: "testRG",
			},
			{
				Name:  "ACRRESOURCEGROUP",
				Value: "testACRRG",
			},
			{
				Name:  "CLUSTERNAME",
				Value: "testCluster",
			},
			{
				Name: "KUSTOMIZEPATH",
				Default: config.BuilderVarDefault{
					Value: "kustomize/overlays/production",
				},
			},
			{
				Name: "MANIFESTPATH",
				Default: config.BuilderVarDefault{
					Value: "manifests",
				},
			},
			{
				Name:  "NAMESPACE",
				Value: "testNamespace",
			},
		},
	}
}
