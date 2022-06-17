package cmd

import (
	"github.com/Azure/draft/pkg/workflows"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

func newGenerateWorkflowCmd() *cobra.Command {

	workflowConfig := &workflows.WorkflowConfig{}
	dest := ""
	var cmd = &cobra.Command{
		Use:   "generate-workflow [flags]",
		Short: "Generates a GitHub Actions workflow for automatic build and deploy to AKS",
		Long: `This command will generate a GitHub Actions workflow to build and deploy an application containerized 
with draft on AKS. This command assumes the 'setup-gh' command has been run properly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowConfig.ValidateAndFillConfig()

			log.Info("--> Generating GitHub Actions workflow")

			if err := workflows.CreateWorkflows(dest, workflowConfig); err != nil {
				return err
			}

			log.Info("Draft has successfully generated a GitHub Actions workflow for your project 😃")

			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&workflowConfig.AksClusterName, "cluster-name", "c", "", "specify the AKS cluster name")
	f.StringVarP(&workflowConfig.AcrName, "registry-name", "r", "", "specify the Azure container registry name")
	f.StringVar(&workflowConfig.ContainerName, "container-name", "", "specify the container image name")
	f.StringVarP(&workflowConfig.ResourceGroupName, "resource-group", "g", "", "Specify the Azure resource group of your AKS cluster")
	f.StringVarP(&dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringVarP(&workflowConfig.BranchName, "branch", "b", "", "specify the GitHub branch to automatically deploy from")

	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}
