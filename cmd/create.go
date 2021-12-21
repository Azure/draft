package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/imiller31/draftv2/pkg/deployments"
	"github.com/imiller31/draftv2/pkg/languages"
	"github.com/imiller31/draftv2/pkg/linguist"
	"github.com/imiller31/draftv2/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ErrNoLanguageDetected is raised when `draft create` does not detect source
// code for linguist to classify, or if there are no packs available for the detected languages.
var ErrNoLanguageDetected = errors.New("no languages were detected")

type createCmd struct {
	appName        string
	lang           string
	dest           string
	repositoryName string
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
			return cc.run()
		},
	}

	f := cmd.Flags()

	f.StringVarP(&cc.appName, "app", "a", "", "name of helm release by default this is randomly generated")
	f.StringVarP(&cc.lang, "lang", "l", "", "the name of the language used to create the k8s deployment")

	return cmd
}

func (cc *createCmd) run() error {
	var err error
	log.Info("detecting language")
	if err = cc.detectLanguage(); err != nil {
		return err
	}

	d := deployments.CreateDeployments(cc.dest)

	selection := &promptui.Select{
		Label: "Select k8s Deployment Type",
		Items: []string{"helm", "kustomize", "manifests"},
	}

	_, deployType, err := selection.Run()
	if err != nil {
		return err
	}

	config := d.GetConfig(deployType)
	customInputs, err := prompts.RunPromptsFromConfig(config)
	if err != nil {
		return err
	}
	log.Infof("--> Creating %s k8s resources", deployType)
	return d.CopyDeploymentFiles(deployType, customInputs)
}

func (cc *createCmd) detectLanguage() error {
	langs, err := linguist.ProcessDir(cc.dest)
	log.Debugf("linguist.ProcessDir(%v) result:\n\nError: %v", cc.dest, err)
	if err != nil {
		return fmt.Errorf("there was an error detecting the language: %s", err)
	}

	hasGo := false
	hasGoMod := false
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

	if len(langs) == 0 {
		return ErrNoLanguageDetected
	}

	supportedLanguages := languages.CreateLanguages(cc.dest)

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

func init() {
	rootCmd.AddCommand(newCreateCmd())
}
