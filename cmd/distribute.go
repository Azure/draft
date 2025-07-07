package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/cmdhelpers"
	"github.com/Azure/draft/pkg/config"
	dryrunpkg "github.com/Azure/draft/pkg/dryrun"
	"github.com/Azure/draft/pkg/handlers"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
)

type distributeCmd struct {
	dest                     string
	provider                 string
	addon                    string
	flagVariables            []string
	templateWriter           templatewriter.TemplateWriter
	templateVariableRecorder config.TemplateVariableRecorder
}

var distributeDryRunRecorder *dryrunpkg.DryRunRecorder

func newDistributeCmd() *cobra.Command {
	dc := &distributeCmd{}
	// distributeCmd represents the distribute command
	var cmd = &cobra.Command{
		Use:   "distribute",
		Short: "Distributes your application resources across Kubernetes clusters using Kubefleet",
		Long: `This command generates Kubefleet ClusterResourcePlacement manifests to distribute your application
		resources across multiple Kubernetes clusters managed by Kubefleet.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := dc.run(); err != nil {
				return err
			}
			log.Info("Draft has successfully created your Kubefleet ClusterResourcePlacement manifest for resource distribution 😃")
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&dc.dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringVarP(&dc.provider, "provider", "p", "azure", "cloud provider")
	f.StringVarP(&dc.addon, "addon", "a", "kubefleet-clusterresourceplacement", "kubefleet addon name")
	f.StringArrayVarP(&dc.flagVariables, "variable", "", []string{}, "pass template variables (e.g. --variable CRP_NAME=demo-crp --variable PLACEMENT_TYPE=PickAll)")

	dc.templateWriter = &writers.LocalFSWriter{}

	return cmd
}

func (dc *distributeCmd) run() error {
	flagVariablesMap = flagVariablesToMap(dc.flagVariables)

	if dryRun {
		distributeDryRunRecorder = dryrunpkg.NewDryRunRecorder()
		dc.templateVariableRecorder = distributeDryRunRecorder
		dc.templateWriter = distributeDryRunRecorder
	}

	updatedDest, err := cmdhelpers.GetAddonDestPath(dc.dest)
	if err != nil {
		log.Errorf("error getting addon destination path: %s", err.Error())
		return err
	}

	// Default to kubefleet-clusterresourceplacement addon, but allow other kubefleet addons
	templateName := "kubefleet-clusterresourceplacement"
	if dc.addon != "" {
		templateName = dc.addon
	}

	// Validate that the addon is a kubefleet addon
	if templateName != "kubefleet-clusterresourceplacement" {
		return fmt.Errorf("distribute command only supports kubefleet addons, got: %s", templateName)
	}

	addonTemplate, err := handlers.GetTemplate(templateName, "", updatedDest, dc.templateWriter)
	if err != nil {
		log.Errorf("error getting kubefleet addon template: %s", err.Error())
		return err
	}
	if addonTemplate == nil {
		return errors.New("DraftConfig is nil")
	}

	addonTemplate.Config.VariableMapToDraftConfig(flagVariablesMap)

	err = cmdhelpers.PromptAddonValues(dc.dest, addonTemplate.Config)
	if err != nil {
		return err
	}

	if dryRun {
		for _, variable := range addonTemplate.Config.Variables {
			dc.templateVariableRecorder.Record(variable.Name, variable.Value)
		}
	}

	err = addonTemplate.Generate()
	if err != nil {
		log.Errorf("error generating kubefleet addon template: %s", err.Error())
		return err
	}

	if dryRun {
		dryRunText, err := json.MarshalIndent(distributeDryRunRecorder.DryRunInfo, "", TWO_SPACES)
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

func init() {
	rootCmd.AddCommand(newDistributeCmd())
}