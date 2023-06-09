package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"

	"github.com/Azure/draft/pkg/reporeader"
	"github.com/Azure/draft/pkg/reporeader/readers"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/deployments"
	dryrunpkg "github.com/Azure/draft/pkg/dryrun"
	"github.com/Azure/draft/pkg/filematches"
	"github.com/Azure/draft/pkg/languages"
	"github.com/Azure/draft/pkg/linguist"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

// ErrNoLanguageDetected is raised when `draft create` does not detect source
// code for linguist to classify, or if there are no packs available for the detected languages.
var ErrNoLanguageDetected = errors.New("no supported languages were detected")
var flagVariablesMap = make(map[string]string)

const LANGUAGE_VARIABLE = "LANGUAGE"
const TWO_SPACES = "  "

// Flag defaults
const emptyDefaultFlagValue = ""
const currentDirDefaultFlagValue = "."

type createCmd struct {
	appName    string
	lang       string
	dest       string
	deployType string

	dockerfileOnly    bool
	deploymentOnly    bool
	skipFileDetection bool
	flagVariables     []string

	createConfigPath string
	createConfig     *CreateConfig

	supportedLangs *languages.Languages

	templateWriter           templatewriter.TemplateWriter
	templateVariableRecorder config.TemplateVariableRecorder
	repoReader               reporeader.RepoReader
}

func newCreateCmd() *cobra.Command {
	cc := &createCmd{}

	cmd := &cobra.Command{
		Use:   "create [flags]",
		Short: "Add minimum required files to the directory",
		Long:  "This command will add the minimum required files to the local directory for your Kubernetes deployment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cc.initConfig(); err != nil {
				return err
			}
			return cc.run()
		},
	}

	f := cmd.Flags()

	f.StringVarP(&cc.createConfigPath, "create-config", "c", emptyDefaultFlagValue, "specify the path to the configuration file")
	f.StringVarP(&cc.appName, "app", "a", emptyDefaultFlagValue, "specify the name of the helm release")
	f.StringVarP(&cc.lang, "language", "l", emptyDefaultFlagValue, "specify the language used to create the Kubernetes deployment")
	f.StringVarP(&cc.dest, "destination", "d", currentDirDefaultFlagValue, "specify the path to the project directory")
	f.StringVarP(&cc.deployType, "deploy-type", "", emptyDefaultFlagValue, "specify deployement type (eg. helm, kustomize, manifests)")
	f.BoolVar(&cc.dockerfileOnly, "dockerfile-only", false, "only create Dockerfile in the project directory")
	f.BoolVar(&cc.deploymentOnly, "deployment-only", false, "only create deployment files in the project directory")
	f.BoolVar(&cc.skipFileDetection, "skip-file-detection", false, "skip file detection step")
	f.StringArrayVarP(&cc.flagVariables, "variable", "", []string{}, "pass additional variables using repeated --variable flag")

	return cmd
}

func (cc *createCmd) initConfig() error {
	if cc.createConfigPath != "" {
		log.Debug("loading config")
		configBytes, err := os.ReadFile(cc.createConfigPath)
		if err != nil {
			return err
		}

		var cfg CreateConfig
		if err = yaml.Unmarshal(configBytes, &cfg); err != nil {
			return err
		}
		cc.createConfig = &cfg
		return nil
	}

	//TODO: create a config for the user and save it for subsequent uses
	cc.createConfig = &CreateConfig{}

	return nil
}

func (cc *createCmd) run() error {
	log.Debugf("config: %s", cc.createConfigPath)

	for _, flagVar := range cc.flagVariables {
		flagVarName, flagVarValue, ok := strings.Cut(flagVar, "=")
		if !ok {
			return fmt.Errorf("invalid variable format: %s", flagVar)
		}
		flagVariablesMap[flagVarName] = flagVarValue
		log.Debugf("flag variable %s=%s", flagVarName, flagVarValue)
	}

	var dryRunRecorder *dryrunpkg.DryRunRecorder
	if dryRun {
		dryRunRecorder = dryrunpkg.NewDryRunRecorder()
		cc.templateVariableRecorder = dryRunRecorder
		cc.templateWriter = dryRunRecorder
	} else {
		cc.templateWriter = &writers.LocalFSWriter{}
	}

	cc.repoReader = &readers.LocalFSReader{}

	detectedLangDraftConfig, languageName, err := cc.detectLanguage()
	if err != nil {
		return err
	}

	err = cc.createFiles(detectedLangDraftConfig, languageName)
	if dryRun {
		cc.templateVariableRecorder.Record(LANGUAGE_VARIABLE, languageName)
		dryRunText, err := json.MarshalIndent(dryRunRecorder.DryRunInfo, "", TWO_SPACES)
		if err != nil {
			return err
		}
		fmt.Println(string(dryRunText))
		if dryRunFile != "" {
			log.Printf("writing dry run info to file %s", dryRunFile)
			err = os.WriteFile(dryRunFile, dryRunText, 0644)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// detectLanguage detects the language used in a project destination directory
// It returns the DraftConfig for that language and the name of the language
func (cc *createCmd) detectLanguage() (*config.DraftConfig, string, error) {
	hasGo := false
	hasGoMod := false
	var langs []*linguist.Language
	var err error
	if cc.createConfig.LanguageType == "" {
		if cc.lang != "" {
			cc.createConfig.LanguageType = cc.lang
		} else {
			log.Info("--- Detecting Language ---")
			langs, err = linguist.ProcessDir(cc.dest)
			log.Debugf("linguist.ProcessDir(%v) result:\n\nError: %v", cc.dest, err)
			if err != nil {
				return nil, "", fmt.Errorf("there was an error detecting the language: %s", err)
			}
			for _, lang := range langs {
				log.Debugf("%s:\t%f (%s)", lang.Language, lang.Percent, lang.Color)
				// For now let's check here for weird stuff like go module support
				if lang.Language == "Go" {
					hasGo = true

					selection := &promptui.Select{
						Label: "Linguist detected Go, do you use Go Modules?",
						Items: []string{"yes", "no"},
					}

					_, selectResponse, err := selection.Run()
					if err != nil {
						return nil, "", err
					}

					hasGoMod = strings.EqualFold(selectResponse, "yes")
				}

				if lang.Language == "Java" {

					selection := &promptui.Select{
						Label: "Linguist detected Java, are you using maven or gradle or gradle wrapper?",
						Items: []string{"gradle", "maven", "gradlew"},
					}

					_, selectResponse, err := selection.Run()
					if err != nil {
						return nil, "", err
					}

					if selectResponse == "gradle" {
						lang.Language = "Gradle"
					} else if selectResponse == "gradlew" {
						lang.Language = "Gradlew"
					}
				}
			}

			log.Debugf("detected %d langs", len(langs))

			if len(langs) == 0 {
				return nil, "", ErrNoLanguageDetected
			}
		}
	}

	cc.supportedLangs = languages.CreateLanguagesFromEmbedFS(template.Dockerfiles, cc.dest)

	if cc.createConfig.LanguageType != "" {
		log.Debug("using configuration language")
		lowerLang := strings.ToLower(cc.createConfig.LanguageType)
		langConfig := cc.supportedLangs.GetConfig(lowerLang)
		if langConfig == nil {
			return nil, "", ErrNoLanguageDetected
		}

		return langConfig, lowerLang, nil
	}

	for _, lang := range langs {
		detectedLang := linguist.Alias(lang)
		log.Infof("--> Draft detected %s (%f%%)\n", detectedLang.Language, detectedLang.Percent)
		lowerLang := strings.ToLower(detectedLang.Language)
		if cc.supportedLangs.ContainsLanguage(lowerLang) {
			if lowerLang == "go" && hasGo && hasGoMod {
				log.Debug("detected go and go module")
				lowerLang = "gomodule"
			}
			langConfig := cc.supportedLangs.GetConfig(lowerLang)
			return langConfig, lowerLang, nil
		}
		log.Infof("--> Could not find a pack for %s. Trying to find the next likely language match...", detectedLang.Language)
	}
	return nil, "", ErrNoLanguageDetected
}

func (cc *createCmd) generateDockerfile(langConfig *config.DraftConfig, lowerLang string) error {
	log.Info("--- Dockerfile Creation ---")
	if cc.supportedLangs == nil {
		return errors.New("supported languages were loaded incorrectly")
	}

	// Extract language-specific defaults from repo
	extractedDefaults, err := cc.supportedLangs.ExtractDefaults(lowerLang, cc.repoReader)
	if err != nil {
		return err
	}
	for _, d := range extractedDefaults {
		langConfig.VariableDefaults = append(langConfig.VariableDefaults, d)
	}

	var inputs map[string]string
	if cc.createConfig.LanguageVariables == nil {
		inputs, err = prompts.RunPromptsFromConfigWithSkips(langConfig, maps.Keys(flagVariablesMap))
		if err != nil {
			return err
		}
	} else {
		inputs, err = validateConfigInputsToPrompts(langConfig.Variables, cc.createConfig.LanguageVariables, langConfig.VariableDefaults)
		if err != nil {
			return err
		}
	}

	if cc.templateVariableRecorder != nil {
		for k, v := range inputs {
			cc.templateVariableRecorder.Record(k, v)
		}
	}

	maps.Copy(inputs, flagVariablesMap)

	if err = cc.supportedLangs.CreateDockerfileForLanguage(lowerLang, inputs, cc.templateWriter); err != nil {
		return fmt.Errorf("there was an error when creating the Dockerfile for language %s: %w", cc.createConfig.LanguageType, err)
	}

	log.Info("--> Creating Dockerfile...\n")
	return err
}

func (cc *createCmd) createDeployment() error {
	log.Info("--- Deployment File Creation ---")
	d := deployments.CreateDeploymentsFromEmbedFS(template.Deployments, cc.dest)
	var deployType string
	var customInputs map[string]string
	var err error

	if cc.createConfig.DeployType != "" {
		deployType = strings.ToLower(cc.createConfig.DeployType)
		deployConfig, err := d.GetConfig(deployType)
		if err != nil {
			return err
		}
		if deployConfig == nil {
			return errors.New("invalid deployment type")
		}
		customInputs, err = validateConfigInputsToPrompts(deployConfig.Variables, cc.createConfig.DeployVariables, deployConfig.VariableDefaults)
		if err != nil {
			return err
		}

	} else {
		if cc.deployType == "" {
			selection := &promptui.Select{
				Label: "Select k8s Deployment Type",
				Items: []string{"helm", "kustomize", "manifests"},
			}

			_, deployType, err = selection.Run()
			if err != nil {
				return err
			}
		} else {
			deployType = cc.deployType
		}

		deployConfig, err := d.GetConfig(deployType)
		if err != nil {
			return err
		}
		customInputs, err = prompts.RunPromptsFromConfigWithSkips(deployConfig, maps.Keys(flagVariablesMap))
		if err != nil {
			return err
		}
	}

	maps.Copy(customInputs, flagVariablesMap)

	if cc.templateVariableRecorder != nil {
		for k, v := range customInputs {
			cc.templateVariableRecorder.Record(k, v)
		}
	}

	log.Infof("--> Creating %s Kubernetes resources...\n", deployType)

	return d.CopyDeploymentFiles(deployType, customInputs, cc.templateWriter)
}

func (cc *createCmd) createFiles(detectedLang *config.DraftConfig, lowerLang string) error {
	// does no further checks without file detection

	if cc.dockerfileOnly && cc.deploymentOnly {
		return errors.New("can only pass in one of --dockerfile-only and --deployment-only")
	}

	cc.detectDefaults(detectedLang, lowerLang)

	if cc.skipFileDetection {
		if !cc.deploymentOnly {
			err := cc.generateDockerfile(detectedLang, lowerLang)
			if err != nil {
				return err
			}
		}
		if !cc.dockerfileOnly {
			err := cc.createDeployment()
			if err != nil {
				return err
			}
		}
		return nil
	}

	// check if the local directory has dockerfile or charts
	hasDockerFile, hasDeploymentFiles, err := filematches.SearchDirectory(cc.dest)
	if err != nil {
		return err
	}

	// prompts user for dockerfile re-creation
	if hasDockerFile && !cc.deploymentOnly {
		selection := &promptui.Select{
			Label: "We found Dockerfile in the directory, would you like to recreate the Dockerfile?",
			Items: []string{"yes", "no"},
		}

		_, selectResponse, err := selection.Run()
		if err != nil {
			return err
		}

		hasDockerFile = strings.EqualFold(selectResponse, "no")
	}

	if cc.deploymentOnly {
		log.Info("--> --deployment-only=true, skipping Dockerfile creation...")
	} else if hasDockerFile {
		log.Info("--> Found Dockerfile in local directory, skipping Dockerfile creation...")
	} else if !cc.deploymentOnly {
		err := cc.generateDockerfile(detectedLang, lowerLang)
		if err != nil {
			return err
		}
	}

	// prompts user for deployment re-creation
	if hasDeploymentFiles && !cc.dockerfileOnly {
		selection := &promptui.Select{
			Label: "We found deployment files in the directory, would you like to create new deployment files?",
			Items: []string{"yes", "no"},
		}

		_, selectResponse, err := selection.Run()
		if err != nil {
			return err
		}

		hasDeploymentFiles = strings.EqualFold(selectResponse, "no")
	}

	if cc.dockerfileOnly {
		log.Info("--> --dockerfile-only=true, skipping deployment file creation...")
	} else if hasDeploymentFiles {
		log.Info("--> Found deployment directory in local directory, skipping deployment file creation...")
	} else if !cc.dockerfileOnly {
		err := cc.createDeployment()
		if err != nil {
			return err
		}
	}

	log.Info("Draft has successfully created deployment resources for your project 😃")
	log.Info("Use 'draft setup-gh' to set up Github OIDC.")

	return nil
}

func (cc *createCmd) detectDefaults(detectedLang *config.DraftConfig, lowerLang string) {
	if lowerLang == "gradlew" || lowerLang == "gradle" {
		// read from build.gradle and detect version
		f, err := os.Open("build.gradle")
		if err != nil {
			log.Warn("Unable to read build.gradle, skipping detection")
			return
		}
		defer f.Close()
		detectedDefaults := make([]config.BuilderVarDefault, 0)
		scanner := bufio.NewScanner(f)
		// this separator is used to split the line from build.gradle ex: sourceCompatibility = '1.8'
		// output will be ['sourceCompatibility', '1.8'] or ["sourceCompatibility", "1.8"]
		separator := func(c rune) bool { return c == ' ' || c == '=' }
		// this func takes care of removing the single or double quotes from split array output
		cutset := func(c rune) bool { return c == '\'' || c == '"' }
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "sourceCompatibility") {
				detectedVersion := strings.FieldsFunc(line, separator)[1] // example array after split ["sourceCompatibility", "1.8"]
				detectedVersion = strings.TrimFunc(detectedVersion, cutset)
				detectedVersion = detectedVersion + "-jre"
				builderVarDefault := config.BuilderVarDefault{
					Name:  "VERSION",
					Value: detectedVersion,
				}
				detectedDefaults = append(detectedDefaults, builderVarDefault)
			}

			if strings.Contains(line, "targetCompatibility") {
				detectedBuilderVersion := strings.FieldsFunc(line, separator)[1] // example array after split ["targetCompatibility", "1.8"]
				detectedBuilderVersion = strings.TrimFunc(detectedBuilderVersion, cutset)
				detectedBuilderVersion = "jdk" + detectedBuilderVersion
				detectedBuilderVar := config.BuilderVarDefault{
					Name:  "BUILDERVERSION",
					Value: detectedBuilderVersion,
				}
				detectedDefaults = append(detectedDefaults, detectedBuilderVar)
			}
			detectedLang.DetectedDefaults = detectedDefaults
		}
	}
}

func init() {
	rootCmd.AddCommand(newCreateCmd())
}

func validateConfigInputsToPrompts(required []config.BuilderVar, provided []UserInputs, defaults []config.BuilderVarDefault) (map[string]string, error) {
	customInputs := make(map[string]string)

	// set inputs to provided values
	for _, variable := range provided {
		customInputs[variable.Name] = variable.Value
	}

	// fill in missing vars using variable default references
	for _, variableDefault := range defaults {
		if customInputs[variableDefault.Name] == "" && variableDefault.ReferenceVar != "" {
			log.Debugf("variable %s is empty, using default referenceVar value from %s", variableDefault.Name, variableDefault.ReferenceVar)
			customInputs[variableDefault.Name] = customInputs[variableDefault.ReferenceVar]
		}
	}

	// fill in missing vars using variable default values
	for _, variableDefault := range defaults {
		if customInputs[variableDefault.Name] == "" && variableDefault.Value != "" {
			log.Debugf("setting default value for %s to %s", variableDefault.Name, variableDefault.Value)
			customInputs[variableDefault.Name] = variableDefault.Value
		}
	}

	for _, variable := range required {
		if _, ok := customInputs[variable.Name]; !ok {
			return nil, fmt.Errorf("config missing required variable: %s with description: %s", variable.Name, variable.Description)
		}
	}

	return customInputs, nil
}
