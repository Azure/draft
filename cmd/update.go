package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/addons"
	"github.com/Azure/draft/pkg/config"
	dryrunpkg "github.com/Azure/draft/pkg/dryrun"
	"github.com/Azure/draft/pkg/templatewriter"
	"github.com/Azure/draft/pkg/templatewriter/writers"
	"github.com/Azure/draft/template"
)

type updateCmd struct {
	dest                     string
	provider                 string
	addon                    string
	flagVariables            []string
	userInputs               map[string]string
	templateWriter           templatewriter.TemplateWriter
	addonFS                  embed.FS
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
	f.StringVarP(&uc.provider, "provider", "p", "azure", "cloud provider")
	f.StringVarP(&uc.addon, "addon", "a", "", "addon name")
	f.StringArrayVarP(&uc.flagVariables, "variable", "", []string{}, "pass a variable non-interactively (ex: --variable foo=bar)")

	uc.templateWriter = &writers.LocalFSWriter{}

	return cmd
}

func (uc *updateCmd) run() error {
	flagVariablesMap := make(map[string]string)
	for _, flagVar := range uc.flagVariables {
		flagVarName, flagVarValue, ok := strings.Cut(flagVar, "=")
		if !ok {
			return fmt.Errorf("invalid variable format: %s", flagVar)
		}
		flagVariablesMap[flagVarName] = flagVarValue
		log.Debugf("flag variable %s=%s", flagVarName, flagVarValue)
	}

	if uc.addon == "" {
		addon, err := addons.PromptAddon(template.Addons, uc.provider)
		if err != nil {
			return err
		}
		uc.addon = addon
	}

	addonConfig, err := addons.GetAddonConfig(template.Addons, uc.provider, uc.addon)
	if err != nil {
		return err
	}

	uc.userInputs, err = addons.PromptAddonValues(uc.dest, flagVariablesMap, addonConfig)
	if err != nil {
		return err
	}
	log.Debugf("addonInputs is: %s", uc.userInputs)

	if dryRun {
		dryRunRecorder = dryrunpkg.NewDryRunRecorder()
		uc.templateVariableRecorder = dryRunRecorder
		uc.templateWriter = dryRunRecorder
		for k, v := range uc.userInputs {
			uc.templateVariableRecorder.Record(k, v)
		}
	}

	err = addons.GenerateAddon(template.Addons, uc.provider, uc.addon, uc.dest, uc.userInputs, uc.templateWriter)

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
