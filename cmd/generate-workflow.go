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
	workflowConfig workflows.WorkflowConfig
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

	for variables := range draftConfig.Variables {
		f.StringVar(&gwCmd.workflowConfig.Variables[variables].Value, draftConfig.Variables[variables].Name, emptyDefaultFlagValue, ""+draftConfig.Variables[variables].Description+DeploymentType)
	}

	f.StringVarP(&gwCmd.workflowConfig.WorkflowName, "workflow-name", "w", emptyDefaultFlagValue, "specify the Github workflow name")
	f.StringVarP(&gwCmd.workflowConfig.BranchName, "branch-name", "b", emptyDefaultFlagValue, "specify the Github branch to automatically deploy from")
	f.StringVar(&gwCmd.workflowConfig.AcrResourceGroup, "acr-resource-group", emptyDefaultFlagValue, "specify the Azure container registry resource group")
	f.StringVarP(&gwCmd.workflowConfig.AcrName, "azure-container-registry", "r", emptyDefaultFlagValue, "specify the Azure container registry name")
	f.StringVar(&gwCmd.workflowConfig.ContainerName, "container-name", emptyDefaultFlagValue, "specify the container image name")
	f.StringVarP(&gwCmd.workflowConfig.ClusterResourceGroup, "cluster-resource-group", "g", emptyDefaultFlagValue, "specify the Azure resource group of your AKS cluster")
	f.StringVarP(&gwCmd.workflowConfig.ClusterName, "cluster-name", "c", emptyDefaultFlagValue, "specify the AKS cluster name")
	f.StringVar(&gwCmd.workflowConfig.Dockerfile, "dockerfile", emptyDefaultFlagValue, "specify the path to the Dockerfile")
	f.StringVarP(&gwCmd.workflowConfig.BuildContextPath, "build-context-path", "x", emptyDefaultFlagValue, "specify the docker build context path")
	f.StringVarP(&gwCmd.workflowConfig.Namespace, "namespace", "n", emptyDefaultFlagValue, "specify the Kubernetes namespace")
	f.StringVar(&gwCmd.workflowConfig.PrivateCluster, "private-cluster", emptyDefaultFlagValue, "specify if the AKS cluster is private")
	f.StringArrayVarP(&gwCmd.flagVariables, "variable", "", []string{}, "pass additional variables")
	gwCmd.templateWriter = &writers.LocalFSWriter{}
	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}

func (gwc *generateWorkflowCmd) generateWorkflows() error {
	var err error

	flagValuesMap := flagVariablesToMap(gwc.flagVariables)

	if gwc.deployType == "" {
		if flagValue := flagValuesMap["deploy-type"]; flagValue == "helm" || flagValue == "kustomize" || flagValue == "manifests" {
			gwc.deployType = flagValuesMap["deploy-type"]
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

	varIdxMap := config.VariableIdxMap(draftConfig.Variables)

	gwc.workflowConfig.SetFlagsToValues(draftConfig, varIdxMap)

	gwc.handleFlagVariables(flagValuesMap, draftConfig, varIdxMap)

	if err = prompts.RunPromptsFromConfigWithSkips(draftConfig, varIdxMap); err != nil {
		return err
	}

	if err := workflows.UpdateProductionDeployments(gwc.deployType, gwc.dest, draftConfig, varIdxMap, gwc.templateWriter); err != nil {
		return fmt.Errorf("update production deployments: %w", err)
	}

	return workflow.CreateWorkflowFiles(gwc.deployType, draftConfig, gwc.templateWriter)
}

func flagVariablesToMap(flagVariables []string) map[string]string {
	flagValuesMap := make(map[string]string)
	for _, flagVar := range flagVariables {
		flagVarName, flagVarValue, ok := strings.Cut(flagVar, "=")
		if !ok {
			log.Fatalf("invalid variable format: %s", flagVar)
		}
		flagValuesMap[flagVarName] = flagVarValue
	}
	return flagValuesMap
}

func (gwc *generateWorkflowCmd) handleFlagVariables(flagValuesMap map[string]string, draftConfig *config.DraftConfig, varIdxMap map[string]int) error {
	for flagVarName, flagVarValue := range flagValuesMap {
		log.Debugf("flag variable %s=%s", flagVarName, flagVarValue)
		switch flagVarName {
		case "destination":
			gwc.dest = flagVarValue
		case "deploy-type":
			continue
		default:
			// comment here
			envArg := strings.ToUpper(strings.ReplaceAll(flagVarName, "-", ""))

			if idx, ok := varIdxMap[envArg]; !ok {
				return fmt.Errorf("flag variable name %s not valid", flagVarName)
			} else {
				draftConfig.Variables[idx].Value = flagVarValue
			}
		}
	}

	return nil
}
