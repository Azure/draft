package azurePipelines

import (
	"fmt"
	"github.com/Azure/draft/pkg/fixtures"
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
		name        string
		deployType  string
		shouldError bool
		setConfig   func(dc *config.DraftConfig)
		cleanUp     func(tempDir string)
	}{
		{
			name:        "kustomize_default_path",
			deployType:  "kustomize",
			shouldError: false,
		},
		{
			name:        "kustomize_given_path",
			deployType:  "kustomize",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("KUSTOMIZEPATH", "kustomize/overlays/production")
			},
		},
		{
			name:        "manifests_default_path",
			deployType:  "manifests",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("PIPELINENAME", "testPipeline")
			},
		},
		{
			name:        "manifests_custom_path",
			deployType:  "manifests",
			shouldError: false,
			setConfig: func(dc *config.DraftConfig) {
				dc.SetVariable("MANIFESTPATH", "test/manifests")
			},
		},
		{
			name:        "invalid",
			deployType:  "invalid",
			shouldError: true,
		},
		{
			name:        "missing_config",
			deployType:  "kustomize",
			shouldError: true,
			setConfig: func(dc *config.DraftConfig) {
				// removing the last variable from draftConfig
				dc.Variables = dc.Variables[:len(dc.Variables)-1]
			},
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

		pipelineFilePath = fmt.Sprintf("%s/.pipelines/%s", tempDir, aksPipelineTemplateFileName)
		if val, ok := draftConfig.FileNameOverrideMap[aksPipelineTemplateFileName]; ok {
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

			// Read the generated content
			generatedContent, err := os.ReadFile(pipelineFilePath)
			assert.Nil(t, err)

			// Validate against the fixture file
			fixturePath := fmt.Sprintf("../fixtures/pipelines/%s.yaml", tt.deployType)
			if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
				t.Fatalf("Fixture file does not exist at path: %s", fixturePath)
			}

			err = fixtures.ValidateContentAgainstFixture(generatedContent, fixturePath)
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
					Value: "test/manifests",
				},
			},
			{
				Name:  "NAMESPACE",
				Value: "testNamespace",
			},
		},
	}
}
