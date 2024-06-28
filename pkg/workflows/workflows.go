package workflows

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/embedutils"
	"github.com/Azure/draft/pkg/osutil"
	"github.com/Azure/draft/pkg/templatewriter"
)

type Workflows struct {
	workflows         map[string]fs.DirEntry
	configs           map[string]*config.DraftConfig
	dest              string
	workflowTemplates fs.FS
}

type DeploymentType string

const (
	parentDirName                          = "workflows"
	configFileName                         = "/draft.yaml"
	emptyDefaultFlagValue                  = ""
	helmDeploymentType      DeploymentType = "helm"
	kustomizeDeploymentType DeploymentType = "kustomize"
	manifestsDeploymentType DeploymentType = "manifests"
)

var allDeploymentTypes = []DeploymentType{helmDeploymentType, kustomizeDeploymentType, manifestsDeploymentType}

func UpdateProductionDeployments(deployType, dest string, draftConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	acr, err := draftConfig.GetVariable("AZURECONTAINERREGISTRY")
	if err != nil {
		return fmt.Errorf("get variable: %w", err)
	}

	containerName, err := draftConfig.GetVariable("CONTAINERNAME")
	if err != nil {
		return fmt.Errorf("get variable: %w", err)
	}

	productionImage := fmt.Sprintf("%s.azurecr.io/%s", acr.Value, containerName.Value)
	switch deployType {
	case "helm":
		return setHelmContainerImage(dest+"/charts/production.yaml", productionImage, templateWriter)
	case "kustomize":
		return setDeploymentContainerImage(dest+"/overlays/production/deployment.yaml", productionImage)
	case "manifests":
		return setDeploymentContainerImage(dest+"/manifests/deployment.yaml", productionImage)
	}
	return nil
}

func setDeploymentContainerImage(filePath, productionImage string) error {

	decode := scheme.Codecs.UniversalDeserializer().Decode
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	k8sObj, _, err := decode(file, nil, nil)
	if err != nil {
		return err
	}
	deploy, ok := k8sObj.(*appsv1.Deployment)
	if !ok {
		return errors.New("could not decode kubernetes deployment")
	}

	if len(deploy.Spec.Template.Spec.Containers) != 1 {
		return errors.New("unsupported number of containers defined in the deployment spec")
	}

	deploy.Spec.Template.Spec.Containers[0].Image = productionImage

	printer := printers.YAMLPrinter{}

	out, err := os.OpenFile(filePath, os.O_RDWR, 0755)
	if err != nil {
		return nil
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Errorf("error closing file: %v", err)
		}
	}()

	return printer.PrintObj(deploy, out)
}

func setHelmContainerImage(filePath, productionImage string, templateWriter templatewriter.TemplateWriter) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var deploy HelmProductionYaml
	err = yaml.Unmarshal(file, &deploy)
	if err != nil {
		return err
	}

	deploy.Image.Repository = productionImage

	out, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	return templateWriter.WriteFile(filePath, out)
}

func (w *Workflows) loadConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := w.workflows[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}

	configPath := path.Join(parentDirName, val.Name(), configFileName)
	configBytes, err := fs.ReadFile(w.workflowTemplates, configPath)
	if err != nil {
		return nil, err
	}

	var draftConfig config.DraftConfig
	if err = yaml.Unmarshal(configBytes, &draftConfig); err != nil {
		return nil, err
	}

	return &draftConfig, nil
}

func (w *Workflows) GetConfig(deployType string) (*config.DraftConfig, error) {
	val, ok := w.configs[deployType]
	if !ok {
		return nil, fmt.Errorf("deploy type %s unsupported", deployType)
	}
	return val, nil
}

func CreateWorkflowsFromEmbedFS(workflowTemplates embed.FS, dest string) *Workflows {
	deployMap, err := embedutils.EmbedFStoMap(workflowTemplates, parentDirName)
	if err != nil {
		log.Fatal(err)
	}

	w := &Workflows{
		workflows:         deployMap,
		dest:              dest,
		configs:           make(map[string]*config.DraftConfig),
		workflowTemplates: workflowTemplates,
	}
	w.populateConfigs()

	return w
}

func (w *Workflows) populateConfigs() {
	for deployType := range w.workflows {
		draftConfig, err := w.loadConfig(deployType)
		if err != nil {
			log.Debugf("no draftConfig found for workflow of deploy type %s", deployType)
			draftConfig = &config.DraftConfig{}
		}
		w.configs[deployType] = draftConfig
	}
}

func (w *Workflows) CreateWorkflowFiles(deployType string, draftConfig *config.DraftConfig, templateWriter templatewriter.TemplateWriter) error {
	val, ok := w.workflows[deployType]
	if !ok {
		return fmt.Errorf("deployment type: %s is not currently supported", deployType)
	}
	srcDir := path.Join(parentDirName, val.Name())
	log.Debugf("source directory for workflow template: %s", srcDir)

	valuesMap, err := draftConfig.VariableMap()
	if err != nil {
		return fmt.Errorf("create workflow files: %w", err)
	}

	draftConfig.ApplyDefaultVariables(valuesMap)

	if err := osutil.CopyDir(w.workflowTemplates, srcDir, w.dest, draftConfig, valuesMap, templateWriter); err != nil {
		return err
	}

	return nil
}

func (w *Workflows) CreateFlags(f *pflag.FlagSet) error {
	type FlagInfo struct {
		description     string
		deploymentTypes []DeploymentType
		isEnvArgCommon  bool
	}

	flags := make(map[string]FlagInfo)
	configs := make(map[DeploymentType]*config.DraftConfig)

	for _, deploymentType := range allDeploymentTypes {
		draftConfig, err := w.GetConfig(string(deploymentType))
		if err != nil {
			return fmt.Errorf("get config: %w", err)
		} else {
			configs[deploymentType] = draftConfig

			for _, variable := range draftConfig.Variables {
				if flag, ok := flags[variable.Name]; ok {
					flag.deploymentTypes = append(flag.deploymentTypes, deploymentType)
					flag.isEnvArgCommon = true
				} else {
					flags[variable.Name] = FlagInfo{
						description:     variable.Description,
						deploymentTypes: []DeploymentType{deploymentType},
						isEnvArgCommon:  false,
					}
				}
			}
		}
	}

	for varName, flagInfo := range flags {
		flagName := strings.ToLower(varName)

		for _, deploymentType := range flagInfo.deploymentTypes {
			variable, err := configs[deploymentType].GetVariable(varName)
			if err != nil {
				return fmt.Errorf("get variable: %w", err)
			}

			if flagInfo.isEnvArgCommon {
				f.StringVar(&variable.Value, flagName, emptyDefaultFlagValue, flagInfo.description)
			} else {
				f.StringVar(&variable.Value, flagName, emptyDefaultFlagValue, fmt.Sprintf("%s (%s)", flagInfo.description, deploymentType))
			}
		}
	}

	return nil
}

func (w *Workflows) HandleFlagVariables(flagValuesMap map[string]string, deploymentType string) error {
	for flagVarName, flagVarValue := range flagValuesMap {
		log.Debugf("flag variable %s=%s", flagVarName, flagVarValue)
		switch flagVarName {
		case "destination":
			w.dest = flagVarValue
		case "deploy-type":
			continue
		default:
			// handles flags that are meant to represent environment arguments
			envArg := strings.ToUpper(flagVarName)

			if variable, err := w.configs[deploymentType].GetVariable(envArg); err != nil {
				return fmt.Errorf("flag variable name %s not valid", flagVarName)
			} else {
				variable.Value = flagVarValue
			}
		}
	}

	return nil
}
