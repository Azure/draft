package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/languages"
	"github.com/Azure/draft/pkg/linguist"
)

func TestRun(t *testing.T) {
	mockCC := &createCmd{}
	mockCC.createConfig = &config.CreateConfig{}
	mockCC.dest = "./.."
	mockCC.createConfig.DeployType = "helm"
	mockCC.createConfig.LanguageVariables = []config.UserInputs{}
	mockCC.createConfig.DeployVariables = []config.UserInputs{}
	mockPortInput := config.UserInputs{Name: "PORT", Value: "8080"}
	mockAppNameInput := config.UserInputs{Name: "APPNAME", Value: "testingCreateCommand"}
	mockCC.createConfig.DeployVariables = append(mockCC.createConfig.DeployVariables, mockPortInput, mockAppNameInput)
	mockCC.createConfig.LanguageVariables = append(mockCC.createConfig.LanguageVariables, mockPortInput)

	oldDockerfile, _ := ioutil.ReadFile("./../Dockerfile")
	oldDockerignore, _ := ioutil.ReadFile("./../.dockerignore")
	
	detectedLang, lowerLang, err := mockCC.mockDetectLanguage()

	assert.False(t, detectedLang == nil)
	assert.False(t, lowerLang == "")
	assert.True(t, err == nil)

	err = mockCC.generateDockerfile(detectedLang, lowerLang)
	assert.True(t, err == nil)

	
	err = mockCC.createDeployment()
	assert.True(t, err == nil)
	err = ioutil.WriteFile("./../Dockerfile", oldDockerfile, 0644)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("./../.dockerignore", oldDockerignore, 0644)
	if err != nil {
		t.Error(err)
	}

	os.RemoveAll("./../charts")
}

func TestInitConfig(t *testing.T) {
	mockCC := &createCmd{}
	mockCC.createConfig = &config.CreateConfig{}
	mockCC.dest = "./.."
	mockCC.createConfigPath = "./../test/templates/config.yaml"

	err := mockCC.initConfig()
	assert.True(t, err == nil)
	assert.True(t, mockCC.createConfig != nil)
}

func (mcc *createCmd) mockDetectLanguage() (*config.DraftConfig, string, error) {
	hasGo := false
	hasGoMod := false
	var langs []*linguist.Language
	var err error

	if mcc.createConfig.LanguageType == "" {
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

	mcc.supportedLangs = languages.CreateLanguages(mcc.dest)

	if mcc.createConfig.LanguageType != "" {
		log.Debug("using configuration language")
		lowerLang := strings.ToLower(mcc.createConfig.LanguageType)
		langConfig := mcc.supportedLangs.GetConfig(lowerLang)
		if langConfig == nil {
			return nil, "", ErrNoLanguageDetected
		}

		return langConfig, lowerLang, nil
	}

	for _, lang := range langs {
		detectedLang := linguist.Alias(lang)
		log.Infof("--> Draft detected %s (%f%%)\n", detectedLang.Language, detectedLang.Percent)
		lowerLang := strings.ToLower(detectedLang.Language)

		if mcc.supportedLangs.ContainsLanguage(lowerLang) {
			if lowerLang == "go" && hasGo && hasGoMod {
				log.Debug("detected go and go module")
				lowerLang = "gomodule"
			}

			langConfig := mcc.supportedLangs.GetConfig(lowerLang)
			return langConfig, lowerLang, nil
		}
		log.Infof("--> Could not find a pack for %s. Trying to find the next likely language match...\n", detectedLang.Language)
	}
	return nil, "", ErrNoLanguageDetected
}