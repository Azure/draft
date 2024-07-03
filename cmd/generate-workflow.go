package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/config"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/pkg/workflows"
	"github.com/Azure/draft/template"
)

type generateWorkflowCmd struct {
	dest           string
	deployType     string
	flagVariables  []string
	templateWriter templatewriter.TemplateWriter
}

func newGenerateWorkflowCmd() *cobra.Command {

	gwCmd := &generateWorkflowCmd{}
	gwCmd.dest = ""
	var cmd = &cobra.Command{
		Use:   "generate-workflow [flags]",
		Short: "Generates a Github workflow for automatic build and deploy to AKS",
		Long: `This command will generate a Github workflow to build and deploy an application containerized 
with draft on AKS. This command assumes the 'setup-gh' command has been run properly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("--> Generating Github workflow")
			if err := gwCmd.generateWorkflows(); err != nil {
				return err
			}

			log.Info("Draft has successfully generated a Github workflow for your project 😃")

			return nil
		},
	}

	f := cmd.Flags()

	f.StringVarP(&gwCmd.dest, "destination", "d", currentDirDefaultFlagValue, "specify the path to the project directory")
	f.StringVarP(&gwCmd.deployType, "deploy-type", "", "", "specify the k8s deployment type (helm, kustomize, manifests)")
	f.StringArrayVarP(&gwCmd.flagVariables, "variable", "", []string{}, "pass environment arguments (e.g. --variable CLUSTERNAME=testCluster --variable DOCKERFILE=./Dockerfile)")
	gwCmd.templateWriter = &writers.LocalFSWriter{}
	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}

func (gwc *generateWorkflowCmd) generateWorkflows() error {
	var err error

	flagVariablesMap = flagVariablesToMap(gwc.flagVariables)

	if gwc.deployType == "" {
		if flagValue := flagVariablesMap["deploy-type"]; flagValue == "helm" || flagValue == "kustomize" || flagValue == "manifests" {
			gwc.deployType = flagVariablesMap["deploy-type"]
		} else {
			selection := &promptui.Select{
				Label: "Select k8s Deployment Type",
				Items: []string{"helm", "kustomize", "manifests"},
			}

			_, gwc.deployType, err = selection.Run()
			if err != nil {
				return err
			}
		}
	}

	workflow := workflows.CreateWorkflowsFromEmbedFS(template.Workflows, gwc.dest)
	draftConfig, err := workflow.GetConfig(gwc.deployType)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	handleFlagVariables(flagVariablesMap, draftConfig, gwc.deployType)

	if err = prompts.RunPromptsFromConfigWithSkips(draftConfig); err != nil {
		return err
	}

	if err := workflows.UpdateProductionDeployments(gwc.deployType, gwc.dest, draftConfig, gwc.templateWriter); err != nil {
		return fmt.Errorf("update production deployments: %w", err)
	}

	return workflow.CreateWorkflowFiles(gwc.deployType, draftConfig, gwc.templateWriter)
}

func handleFlagVariables(flagVariablesMap map[string]string, draftConfig *config.DraftConfig, infoType string) {
	for flagName, flagValue := range flagVariablesMap {
		log.Debugf("flag variable %s=%s", flagName, flagValue)
		// handles flags that are meant to represent environment arguments
		if variable, err := draftConfig.GetVariable(flagName); err != nil {
			log.Debugf("flag variable name %s not a valid environment argument for %s", flagName, infoType)
			continue
		} else {
			variable.Value = flagValue
		}
	}
}

func flagVariablesToMap(flagVariables []string) map[string]string {
	flagVariablesMap := make(map[string]string)
	for _, flagVar := range flagVariables {
		flagVarName, flagVarValue, ok := strings.Cut(flagVar, "=")
		if !ok {
			log.Fatalf("invalid variable format: %s", flagVar)
		}
		flagVariablesMap[flagVarName] = flagVarValue
	}
	return flagVariablesMap
}
