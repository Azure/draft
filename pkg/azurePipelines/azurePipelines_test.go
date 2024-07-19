package azurePipelines

import (
	"os"
	"testing"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
	"github.com/stretchr/testify/assert"
)

func TestCreatePipelines(t *testing.T) {
	dest := "."
	templateWriter := &writers.LocalFSWriter{}

	tests := []struct {
		name             string
		deployType       string
		shouldError      bool
		setConfig        func(dc *config.DraftConfig)
		pipelineFilePath string
		cleanUp          func()
	}{
		{
			name:             "kustomize_default_path",
			deployType:       "kustomize",
			shouldError:      false,
			pipelineFilePath: ".pipelines/azure-kubernetes-service-kustomize.yaml",
			cleanUp: func() {
				err := os.RemoveAll(".pipelines")
				assert.Nil(t, err)
			},
		},
		{
			name:        "kustomize_given_path",
			deployType:  "kustomize",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("KUSTOMIZEPATH", "test/kustomize/overlays/production")
			},
			pipelineFilePath: ".pipelines/azure-kubernetes-service-kustomize.yaml",
			cleanUp: func() {
				err := os.RemoveAll(".pipelines")
				assert.Nil(t, err)
			},
		},
		{
			name:             "manifests_default_path",
			deployType:       "manifests",
			shouldError:      false,
			pipelineFilePath: ".pipelines/azure-kubernetes-service.yaml",
			cleanUp: func() {
				err := os.RemoveAll(".pipelines")
				assert.Nil(t, err)
			},
		},
		{
			name:        "manifests_custom_path",
			deployType:  "manifests",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("MANIFESTPATH", "test/manifests")
			},
			pipelineFilePath: ".pipelines/azure-kubernetes-service.yaml",
			cleanUp: func() {
				err := os.RemoveAll(".pipelines")
				assert.Nil(t, err)
			},
		},
		{
			name:        "invalid",
			deployType:  "invalid",
			shouldError: true,
			cleanUp: func() {
			},
		},
		{
			name:        "missing_config",
			deployType:  "kustomize",
			shouldError: true,
			setConfig: func(dc *config.DraftConfig) {
				// removing the last variable from draftConfig
				dc.Variables = dc.Variables[:len(dc.Variables)-1]
			},
			pipelineFilePath: ".pipelines/azure-kubernetes-service-kustomize.yaml",
			cleanUp: func() {
				err := os.RemoveAll(".pipelines")
				assert.Nil(t, err)
			},
		},
	}

	for _, tt := range tests {
		draftConfig := newDraftConfig()

		pipelines, err := CreatePipelinesFromEmbedFS(template.AzurePipelines, dest)
		assert.Nil(t, err)

		if tt.setConfig != nil {
			tt.setConfig(draftConfig)
		}

		err = pipelines.CreatePipelineFiles(tt.deployType, draftConfig, templateWriter)

		if tt.shouldError {
			assert.NotNil(t, err)
			_, err = os.Stat(tt.pipelineFilePath)
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			_, err = os.Stat(tt.pipelineFilePath)
			assert.Nil(t, err)
		}
		tt.cleanUp()
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
