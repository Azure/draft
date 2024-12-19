package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/cmdhelpers"
	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/prompts"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
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
	f.StringArrayVarP(&gwCmd.flagVariables, "variable", "", []string{}, "pass template variables (e.g. --variable CLUSTERNAME=testCluster --variable DOCKERFILE=./Dockerfile)")
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
		selection := &promptui.Select{
			Label: "Select k8s Deployment Type",
			Items: []string{"helm", "kustomize", "manifests"},
		}

		_, gwc.deployType, err = selection.Run()
		if err != nil {
			return err
		}
	}

	t, err := handlers.GetTemplate(fmt.Sprintf("github-workflow-%s", gwc.deployType), "", gwc.dest, gwc.templateWriter)
	if err != nil {
		return fmt.Errorf("failed to get template: %e", err)
	}
	if t == nil {
		return fmt.Errorf("template is nil")
	}

	t.Config.VariableMapToDraftConfig(flagVariablesMap)

	if err = prompts.RunPromptsFromConfigWithSkips(t.Config); err != nil {
		return err
	}

	if err := cmdhelpers.UpdateProductionDeployments(gwc.deployType, gwc.dest, t.Config, gwc.templateWriter); err != nil {
		return fmt.Errorf("update production deployments: %w", err)
	}

	return t.Generate()
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
