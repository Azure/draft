package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/imiller31/draftv2/pkg/configs"
	"github.com/imiller31/draftv2/pkg/deployments"
	"github.com/imiller31/draftv2/pkg/languages"
	"github.com/imiller31/draftv2/pkg/linguist"
	"github.com/imiller31/draftv2/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ErrNoLanguageDetected is raised when `draft create` does not detect source
// code for linguist to classify, or if there are no packs available for the detected languages.
var ErrNoLanguageDetected = errors.New("no languages were detected")

type createCmd struct {
	appName        string
	lang           string
	dest           string
	repositoryName string

	createConfigPath string
	createConfig     *configs.CreateConfig
}

func newCreateCmd() *cobra.Command {
	cc := &createCmd{}

	cmd := &cobra.Command{
		Use:   "create [path]",
		Short: "add minimum viable files to deploy to k8s",
		Long:  "This command will add the necessary files to the local directory for deployment to k8s",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Debugf("number of args passed: %d", len(args))
			if len(args) > 0 {
				cc.dest = args[0]
			} else {
				cc.dest = "."
			}

			cc.initConfig()
			return cc.run()
		},
	}

	f := cmd.Flags()

	f.StringVarP(&cc.createConfigPath, "createConfig", "c", "", "will use configuration given if set")
	f.StringVarP(&cc.appName, "app", "a", "", "name of helm release by default this is randomly generated")
	f.StringVarP(&cc.lang, "lang", "l", "", "the name of the language used to create the k8s deployment")

	return cmd
}

func (cc *createCmd) initConfig() error {
	if cc.createConfigPath != "" {
		log.Debug("loading config")
		configBytes, err := os.ReadFile(cc.createConfigPath)
		if err != nil {
			return err
		}

		viper.SetConfigFile("yaml")
		if err = viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return err
		}
		var cfg configs.CreateConfig
		if err = viper.Unmarshal(&cfg); err != nil {
			return err
		}

		cc.createConfig = &cfg
		return nil
	}

	//TODO: create a config for the user and save it for subsequent uses
	cc.createConfig = &configs.CreateConfig{}

	return nil
}

func (cc *createCmd) run() error {
	log.Debugf("config: %s", cc.createConfigPath)
	var err error
	log.Info("detecting language")
	if err = cc.detectLanguage(); err != nil {
		return err
	}

	return cc.createDeployment()
}

func (cc *createCmd) detectLanguage() error {
	hasGo := false
	hasGoMod := false
	var langs []*linguist.Language
	var err error
	if cc.createConfig.LanguageType == "" {
		langs, err = linguist.ProcessDir(cc.dest)
		log.Debugf("linguist.ProcessDir(%v) result:\n\nError: %v", cc.dest, err)
		if err != nil {
			return fmt.Errorf("there was an error detecting the language: %s", err)
		}
		for _, lang := range langs {
			log.Debugf("%s:\t%f (%s)", lang.Language, lang.Percent, lang.Color)
			// For now let's check here for weird stuff like go module support
			if lang.Language == "Go" {
				hasGo = true
			}
			if lang.Language == "Go Module" {
				hasGoMod = true
			}
		}

		log.Debugf("detected %d langs", len(langs))

		if len(langs) == 0 {
			return ErrNoLanguageDetected
		}
	}

	supportedLanguages := languages.CreateLanguages(cc.dest)

	if cc.createConfig.LanguageType != "" {
		log.Debug("using configuration language")
		lowerLang := strings.ToLower(cc.createConfig.LanguageType)
		langConfig := supportedLanguages.GetConfig(lowerLang)
		if langConfig == nil {
			return ErrNoLanguageDetected
		}
		inputs, err := validateConfigInputsToPrompts(langConfig.Variables, cc.createConfig.LanguageVariables)
		if err != nil {
			return err
		}

		if err = supportedLanguages.CreateDockerfileForLanguage(lowerLang, inputs); err != nil {
			return fmt.Errorf("there was an error when creating the Dockerfile for language %s: %w", cc.createConfig.LanguageType, err)
		}

		return nil
	}

	for _, lang := range langs {
		detectedLang := linguist.Alias(lang)
		log.Infof("--> Draft detected %s (%f%%)\n", detectedLang.Language, detectedLang.Percent)
		lowerLang := strings.ToLower(detectedLang.Language)
		if supportedLanguages.ContainsLanguage(lowerLang) {
			if lowerLang == "go" && hasGo && hasGoMod {
				log.Debug("detected go and go module")
				lowerLang = "gomodule"
			}
			langConfig := supportedLanguages.GetConfig(lowerLang)
			inputs, err := prompts.RunPromptsFromConfig(langConfig)
			if err != nil {
				return err
			}

			if err = supportedLanguages.CreateDockerfileForLanguage(lowerLang, inputs); err != nil {
				return fmt.Errorf("there was an error when creating the Dockerfile for language %s: %w", detectedLang.Language, err)
			}
			return err
		}
		log.Infof("--> Could not find a pack for %s. Trying to find the next likely language match...\n", detectedLang.Language)
	}
	return ErrNoLanguageDetected
}

func (cc *createCmd) createDeployment() error {
	d := deployments.CreateDeployments(cc.dest)
	var deployType string
	var customInputs map[string]string
	var err error
	if cc.createConfig.DeployType != "" {
		deployType = strings.ToLower(cc.createConfig.DeployType)
		config := d.GetConfig(deployType)
		if config == nil {
			return errors.New("invalid deployment type")
		}
		customInputs, err = validateConfigInputsToPrompts(config.Variables, cc.createConfig.DeployVariables)
		if err != nil {
			return err
		}

	} else {
		selection := &promptui.Select{
			Label: "Select k8s Deployment Type",
			Items: []string{"helm", "kustomize", "manifests"},
		}

		_, deployType, err := selection.Run()
		if err != nil {
			return err
		}

		config := d.GetConfig(deployType)
		customInputs, err = prompts.RunPromptsFromConfig(config)
		if err != nil {
			return err
		}
	}

	log.Infof("--> Creating %s k8s resources", deployType)
	return d.CopyDeploymentFiles(deployType, customInputs)
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
}

func validateConfigInputsToPrompts(required []configs.BuilderVar, provided []configs.UserInputs) (map[string]string, error) {
	customInputs := make(map[string]string)
	for _, variable := range provided {
		customInputs[variable.Name] = variable.Value
	}

	for _, variable := range required {
		if _, ok := customInputs[variable.Name]; !ok {
			return nil, errors.New(fmt.Sprintf("config missing language variable: %s with description: %s", variable.Name, variable.Description))
		}
	}

	return customInputs, nil
}
