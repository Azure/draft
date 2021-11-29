/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/imiller31/draftv2/pkg/deployments"
	"github.com/imiller31/draftv2/pkg/linguist"
	"strings"

	"github.com/imiller31/draftv2/pkg/languages"
	"github.com/imiller31/draftv2/pkg/prompts"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ErrNoLanguageDetected is raised when `draft create` does not detect source
// code for linguist to classify, or if there are no packs available for the detected languages.
var ErrNoLanguageDetected = errors.New("no languages were detected")

type createCmd struct {
	appName string
	lang string
	deployType string
	dest string
	repositoryName string
}

func newCreateCmd() *cobra.Command {
	cc := &createCmd{}

	cmd := &cobra.Command{
		Use: "create [path]",
		Short: "add minimum viable files to deploy to k8s",
		Long: "This command will add the necessary files to the local directory for deployment to k8s",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cc.dest = args[0]
			}
			return cc.run();
			},
	}

	f := cmd.Flags()

	f.StringVarP(&cc.appName, "app", "a", "", "name of helm release by default this is randomly generated")
	f.StringVarP(&cc.deployType, "deploy-type", "d", "helm", "deployment type (defaults to helm): helm, kustomize, manifest")
	f.StringVarP(&cc.lang, "lang", "l", "", "the name of the language used to create the k8s deployment")

	return cmd
}

func (cc *createCmd) run() error {
	var err error
	log.Info("detecting language")
	if err = cc.detectLanguage(); err != nil {
		return err
	}

	d := deployments.CreateDeployments()

	switch cc.deployType {
		case "helm": err = cc.createHelm(d)

		case "kustomize": err = cc.createKustomize(d)

		case "manifest": err = cc.createManifest(d)
	}

	return err
}

func (cc *createCmd) detectLanguage() error {
	langs, err := linguist.ProcessDir(".")
	log.Debugf("linguist.ProcessDir('.') result:\n\nError: %v", err)
	if err != nil {
		return fmt.Errorf("there was an error detecting the language: %s", err)
	}

	for _, lang := range langs {
		log.Debugf("%s:\t%f (%s)", lang.Language, lang.Percent, lang.Color)
	}

	if len(langs) == 0 {
		return ErrNoLanguageDetected
	}

	supportedLanguages := languages.CreateLanguages()

	for _, lang := range langs {
		detectedLang := linguist.Alias(lang)
		log.Infof("--> Draft detected %s (%f%%)\n", detectedLang.Language, detectedLang.Percent)
		lowerLang := strings.ToLower(detectedLang.Language)
		if supportedLanguages.ContainsLanguage(lowerLang) {

			langConfig := supportedLanguages.GetConfig(lowerLang)
			templatePrompts := make([]*prompts.TemplatePrompt, 0)

			for _, customPrompt := range langConfig.Variables {
				prompt := &promptui.Prompt{
					Label: customPrompt.Description,
					Validate: func(s string) error {
						if len(s) <= 0 {
							return fmt.Errorf("input must be greater than 0")
						}
						return nil
					},
				}
				templatePrompts = append(templatePrompts, &prompts.TemplatePrompt{
					Prompt: prompt,
					OverrideString: customPrompt.Name,
				})
			}

			inputs := make(map[string]string)

			for _, prompt := range templatePrompts {
				input, err := prompt.Prompt.Run()
				if err != nil {
					return err
				}
				inputs[prompt.OverrideString] = input
			}

			if err = supportedLanguages.CreateDockerfileForLanguage(lowerLang, inputs); err != nil {
				return fmt.Errorf("there was an error when creating the Dockerfile for language %s: %w", detectedLang.Language, err)
			}
			return err
		}
		log.Infof( "--> Could not find a pack for %s. Trying to find the next likely language match...\n", detectedLang.Language)
	}
	return ErrNoLanguageDetected
}

func (cc *createCmd) createHelm(d *deployments.Deployments) error {
	log.Info("--> Creating helm chart")
	return d.CopyDeploymentFiles("helm")
}

func (cc *createCmd) createKustomize(d *deployments.Deployments) error {
	return errors.New("kustomize not yet implemented")
}

func (cc *createCmd) createManifest(d *deployments.Deployments) error {
	return errors.New("manifests not yet implemented")
}

func init() {
	rootCmd.AddCommand(newCreateCmd())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
