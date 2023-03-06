package cmd

import (
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/pkg/workflows"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

type generateWorkflowCmd struct {
	workflowConfig workflows.WorkflowConfig
	dest           string
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
			gwCmd.workflowConfig.ValidateAndFillConfig()

			log.Info("--> Generating Github workflow")

			if err := workflows.CreateWorkflows(gwCmd.dest, &gwCmd.workflowConfig, gwCmd.flagVariables, gwCmd.templateWriter); err != nil {
				return err
			}

			log.Info("Draft has successfully generated a Github workflow for your project 😃")

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&gwCmd.workflowConfig.AksClusterName, "cluster-name", "c", "", "specify the AKS cluster name")
	f.StringVarP(&gwCmd.workflowConfig.AcrName, "registry-name", "r", "", "specify the Azure container registry name")
	f.StringVar(&gwCmd.workflowConfig.ContainerName, "container-name", "", "specify the container image name")
	f.StringVarP(&gwCmd.workflowConfig.ResourceGroupName, "resource-group", "g", "", "specify the Azure resource group of your AKS cluster")
	f.StringVarP(&gwCmd.dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringVarP(&gwCmd.workflowConfig.BranchName, "branch", "b", "", "specify the Github branch to automatically deploy from")
	f.StringArrayVarP(&gwCmd.flagVariables, "variable", "", []string{}, "add additional variables in-line using --variable flag")
	gwCmd.templateWriter = &writers.LocalFSWriter{}
	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}
