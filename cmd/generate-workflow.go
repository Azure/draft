package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/pkg/workflows"
)

type generateWorkflowCmd struct {
	workflowConfig workflows.WorkflowConfig
	dest           string
	deployType     string
	flagVariables  []string
	templateWriter templatewriter.TemplateWriter
}

var flagValuesMap map[string]string

func newGenerateWorkflowCmd() *cobra.Command {

	gwCmd := &generateWorkflowCmd{}
	gwCmd.dest = ""
	var cmd = &cobra.Command{
		Use:   "generate-workflow [flags]",
		Short: "Generates a Github workflow for automatic build and deploy to AKS",
		Long: `This command will generate a Github workflow to build and deploy an application containerized 
with draft on AKS. This command assumes the 'setup-gh' command has been run properly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() != 0 {
				flagValuesMap = gwCmd.workflowConfig.SetFlagValuesToMap()
			}
			log.Info("--> Generating Github workflow")
			if err := workflows.CreateWorkflows(gwCmd.dest, gwCmd.deployType, gwCmd.flagVariables, gwCmd.templateWriter, flagValuesMap); err != nil {
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
	f.StringVarP(&gwCmd.workflowConfig.ResourceGroupName, "resource-group", "g", "", "Specify the Azure resource group of your AKS cluster")
	f.StringVarP(&gwCmd.dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringVarP(&gwCmd.workflowConfig.BranchName, "branch", "b", "", "specify the Github branch to automatically deploy from")
	f.StringVar(&gwCmd.deployType, "deploy-type", "", "specify the type of deployment")
	f.StringArrayVarP(&gwCmd.flagVariables, "variable", "", []string{}, "pass additional variables")
	f.StringVarP(&gwCmd.workflowConfig.BuildContextPath, "build-context-path", "x", "", "specify the docker build context path")
	gwCmd.templateWriter = &writers.LocalFSWriter{}
	return cmd
}

func init() {
	rootCmd.AddCommand(newGenerateWorkflowCmd())
}
