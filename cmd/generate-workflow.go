package cmd

import (
	"github.com/Azure/draftv2/pkg/workflows"
	"github.com/spf13/cobra"
)

func newGenerateWorkflowCmd() *cobra.Command {

	workflowConfig := &workflows.WorkflowConfig{}
	dest := ""
	var cmd = &cobra.Command{
		Use:   "generate-workflow",
		Short: "generates a github workflow for automatic build and deploy to AKS",
		Long:  `This command will generate a github workflow to build and deploy an application containerized with draft. This command assumes the 'setup-gh' command has been run properly.'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowConfig.ValidateAndFillConfig()

			return workflows.CreateWorkflows(dest, workflowConfig)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&workflowConfig.AksClusterName, "cluster-name", "c", "", "name of AKS cluster")
	f.StringVarP(&workflowConfig.AcrName, "registry-name", "r", "", "the Azure container registry name")
	f.StringVar(&workflowConfig.ContainerName, "container-name", "", "the name of the container image")
	f.StringVarP(&workflowConfig.ResourceGroupName, "resource-group", "g", "", "the Azure resource group of your AKS cluster")
	f.StringVarP(&dest, "destination", "d", ".", "root of repository for gh workflow")

	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}
