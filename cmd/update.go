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

type updateCmd struct {
	dest                     string
	flagVariables            []string
	templateWriter           templatewriter.TemplateWriter
	templateVariableRecorder config.TemplateVariableRecorder
}

var dryRunRecorder *dryrunpkg.DryRunRecorder

func newUpdateCmd() *cobra.Command {
	uc := &updateCmd{}
	// updateCmd represents the update command
	var cmd = &cobra.Command{
		Use:   "update",
		Short: "Updates your application to be internet accessible",
		Long: `This command automatically updates your yaml files as necessary so that your application
		will be able to receive external requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := uc.run(); err != nil {
				return err
			}
			log.Info("Draft has successfully updated your yaml files so that your application will be able to receive external requests 😃")
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&uc.dest, "destination", "d", ".", "specify the path to the project directory")
	f.StringArrayVarP(&uc.flagVariables, "variable", "", []string{}, "pass template variables (e.g. --variable ingress-tls-cert-keyvault-uri=test.uri ingress-host=host)")

	uc.templateWriter = &writers.LocalFSWriter{}

	return cmd
}

func (uc *updateCmd) run() error {
	flagVariablesMap = flagVariablesToMap(uc.flagVariables)

	if dryRun {
		dryRunRecorder = dryrunpkg.NewDryRunRecorder()
		uc.templateVariableRecorder = dryRunRecorder
		uc.templateWriter = dryRunRecorder
	}

	ingressTemplate, err := handlers.GetTemplate("app-routing-ingress", "", uc.dest, uc.templateWriter)
	if err != nil {
		log.Errorf("error getting ingress template: %s", err.Error())
		return err
	}
	if ingressTemplate == nil {
		return errors.New("DraftConfig is nil")
	}

	ingressTemplate.Config.VariableMapToDraftConfig(flagVariablesMap)

	err = cmdhelpers.PromptAddonValues(uc.dest, ingressTemplate.Config)
	if err != nil {
		return err
	}

	if dryRun {
		for _, variable := range ingressTemplate.Config.Variables {
			uc.templateVariableRecorder.Record(variable.Name, variable.Value)
		}
	}

	err = ingressTemplate.Generate()
	if err != nil {
		log.Errorf("error generating ingress template: %s", err.Error())
		return err
	}

	if dryRun {
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

func init() {
	rootCmd.AddCommand(newUpdateCmd())
}
