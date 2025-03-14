package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/linguist"
	"github.com/Azure/draft/pkg/reporeader"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

func TestRun(t *testing.T) {
	deployTypes := []string{"helm", "kustomize", "manifests"}

	for _, deployType := range deployTypes {
		var testDir = t.TempDir()

		testCreateConfig := CreateConfig{
			DeployType:        deployType,
			LanguageVariables: []UserInputs{{Name: "PORT", Value: "8080"}},
			DeployVariables: []UserInputs{
				{Name: "PORT", Value: "8080"},
				{Name: "APPNAME", Value: "testingCreateCommand"},
				{Name: "DOCKERFILENAME", Value: "Dockerfile"},
			},
		}
		flagVariablesMap = map[string]string{"PORT": "8080", "APPNAME": "testingCreateCommand", "VERSION": "1.18", "SERVICEPORT": "8080", "NAMESPACE": "testNamespace", "IMAGENAME": "testImage", "IMAGETAG": "latest", "DOCKERFILENAME": "test.Dockerfile"}
		mockCC := createCmd{
			deployType:     deployType,
			dest:           testDir,
			createConfig:   &testCreateConfig,
			templateWriter: &writers.LocalFSWriter{},
		}

		err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("//placeholder"), 0644)
		assert.Nil(t, err)
		detectedLang, lowerLang, err := mockCC.mockDetectLanguage()
		assert.NotNil(t, detectedLang)
		assert.False(t, lowerLang == "")
		assert.Nil(t, err)

		err = mockCC.generateDockerfile(detectedLang, lowerLang)
		assert.Nil(t, err)

		//when language variables are passed in --variable flag
		mockCC.createConfig.LanguageVariables = nil
		mockCC.lang = "go"
		detectedLang, lowerLang, err = mockCC.mockDetectLanguage()
		assert.NotNil(t, detectedLang)
		assert.False(t, lowerLang == "")
		assert.Nil(t, err)
		err = mockCC.generateDockerfile(detectedLang, lowerLang)
		assert.Nil(t, err)

		//deployment variables passed through --variable flag
		err = mockCC.generateDeployment()
		assert.Nil(t, err)
		//check if deployment files have been created
		deploymentFiles, err := getAllDeploymentFiles(filepath.Join("../template/deployments", mockCC.deployType))
		assert.Nil(t, err)
		for _, fileName := range deploymentFiles {
			_, err = os.Stat(filepath.Join(testDir, fileName))
			assert.Nil(t, err)
		}

		//deployment variables passed through createConfig
		err = mockCC.generateDeployment()
		assert.Nil(t, err)
		//check if deployment files have been created
		deploymentFiles, err = getAllDeploymentFiles(path.Join("../template/deployments", mockCC.createConfig.DeployType))
		assert.Nil(t, err)
		for _, fileName := range deploymentFiles {
			_, err = os.Stat(filepath.Join(testDir, fileName))
			assert.Nil(t, err)
		}
	}
}

func TestRunCreateDockerfileWithRepoReader(t *testing.T) {

	testRepoReader := &reporeader.FakeRepoReader{Files: map[string][]byte{
		"foo.py":  []byte("print('Hello World')"),
		"main.py": []byte("print('Hello World')"),
	}}

	testCreateConfig := CreateConfig{LanguageType: "python", LanguageVariables: []UserInputs{{Name: "PORT", Value: "8080"}}}
	mockCC := createCmd{createConfig: &testCreateConfig, repoReader: testRepoReader, templateWriter: &writers.LocalFSWriter{}}

	detectedLang, lowerLang, err := mockCC.mockDetectLanguage()
	assert.NotNil(t, detectedLang)
	assert.True(t, lowerLang == "python")
	assert.Nil(t, err)

	err = mockCC.generateDockerfile(detectedLang, lowerLang)
	assert.Nil(t, err)

	dockerFileContent, err := ioutil.ReadFile("Dockerfile")
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, string(dockerFileContent), "CMD [\"main.py\"]")

	err = os.Remove("Dockerfile")
	if err != nil {
		t.Error(err)
	}
	err = os.RemoveAll(".dockerignore")
	if err != nil {
		t.Error(err)
	}
}

func TestInitConfig(t *testing.T) {
	mockCC := &createCmd{}
	mockCC.createConfig = &CreateConfig{}
	mockCC.dest = "./.."
	mockCC.createConfigPath = "./../test/templates/config.yaml"

	err := mockCC.initConfig()
	assert.Nil(t, err)
	assert.NotNil(t, mockCC.createConfig)
}

func TestValidateConfigInputsToPromptsPass(t *testing.T) {
	required := config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name: "REQUIRED_PROVIDED",
			},
			{
				Name: "REQUIRED_DEFAULTED",
				Default: config.BuilderVarDefault{
					Value: "DEFAULT_VALUE",
				},
			},
		},
	}
	provided := []UserInputs{
		{Name: "REQUIRED_PROVIDED", Value: "PROVIDED_VALUE"},
	}

	err := validateConfigInputsToPrompts(&required, provided)
	assert.Nil(t, err)

	err = required.ApplyDefaultVariables()
	assert.Nil(t, err)

	var1, err := required.GetVariable("REQUIRED_PROVIDED")
	assert.Nil(t, err)
	assert.Equal(t, var1.Value, "PROVIDED_VALUE")

	var2, err := required.GetVariable("REQUIRED_DEFAULTED")
	assert.Nil(t, err)
	assert.Equal(t, var2.Value, "DEFAULT_VALUE")
}

func TestValidateConfigInputsToPromptsMissing(t *testing.T) {
	required := config.DraftConfig{
		Variables: []*config.BuilderVar{
			{
				Name: "REQUIRED_PROVIDED",
			},
			{
				Name: "REQUIRED_MISSING",
			},
		},
	}
	provided := []UserInputs{
		{Name: "REQUIRED_PROVIDED"},
	}

	err := validateConfigInputsToPrompts(&required, provided)
	assert.Nil(t, err)

	err = required.ApplyDefaultVariables()
	assert.NotNil(t, err)
}

func (mcc *createCmd) mockDetectLanguage() (*handlers.Template, string, error) {
	hasGo := false
	hasGoMod := false
	var langs []*linguist.Language
	var err error

	if mcc.createConfig.LanguageType == "" {
		if mcc.lang != "" {
			mcc.createConfig.LanguageType = mcc.lang
		} else {
			langs, err = linguist.ProcessDir(mcc.dest)
			log.Debugf("linguist.ProcessDir(%v) result:\n\nError: %v", mcc.dest, err)
			if err != nil {
				return nil, "", fmt.Errorf("there was an error detecting the language: %s", err)
			}

			for _, lang := range langs {
				log.Debugf("%s:\t%f (%s)", lang.Language, lang.Percent, lang.Color)
			}

			log.Debugf("detected %d langs", len(langs))

			if len(langs) == 0 {
				return nil, "", ErrNoLanguageDetected
			}
		}
	}

	if mcc.createConfig.LanguageType != "" {
		log.Debug("using configuration language")
		lowerLang := strings.ToLower(mcc.createConfig.LanguageType)
		langConfig, err := handlers.GetTemplate(fmt.Sprintf("dockerfile-%s", lowerLang), "", mcc.dest, mcc.templateWriter)
		if err != nil {
			return nil, "", err
		}
		if langConfig == nil {
			return nil, "", ErrNoLanguageDetected
		}

		return langConfig, lowerLang, nil
	}

	for _, lang := range langs {
		detectedLang := linguist.Alias(lang)
		log.Infof("--> Draft detected %s (%f%%)\n", detectedLang.Language, detectedLang.Percent)
		lowerLang := strings.ToLower(detectedLang.Language)

		if handlers.IsValidTemplate(fmt.Sprintf("dockerfile-%s", lowerLang)) {
			if lowerLang == "go" && hasGo && hasGoMod {
				log.Debug("detected go and go module")
				lowerLang = "gomodule"
			}

			langConfig, err := handlers.GetTemplate(fmt.Sprintf("dockerfile-%s", lowerLang), "", mcc.dest, mcc.templateWriter)
			if err != nil {
				return nil, "", err
			}
			if langConfig == nil {
				return nil, "", ErrNoLanguageDetected
			}
			return langConfig, lowerLang, nil
		}
		log.Infof("--> Could not find a pack for %s. Trying to find the next likely language match...\n", detectedLang.Language)
	}
	return nil, "", ErrNoLanguageDetected
}

func TestDefaultValues(t *testing.T) {
	assert.Equal(t, emptyDefaultFlagValue, "")
	assert.Equal(t, currentDirDefaultFlagValue, ".")
}

func getAllDeploymentFiles(src string) ([]string, error) {
	deploymentFiles := []string{}
	err := filepath.Walk(src,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			filePath := strings.ReplaceAll(path, src, "")
			if info.Name() != "draft.yaml" {
				deploymentFiles = append(deploymentFiles, filePath)
			}
			return nil
		})
	return deploymentFiles, err
}
