package cmd

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/deployments"
	"github.com/Azure/draft/pkg/languages"
	"github.com/Azure/draft/template"
)

type Format string

const (
	JSON Format = "json"
)

type infoCmd struct {
	format string
	info   *draftInfo
}

type draftInfo struct {
	SupportedLanguages       []string `json:"supported_languages"`
	SupportedDeploymentTypes []string `json:"supported_deployment_types"`
}

func newInfoCmd() *cobra.Command {
	ic := &infoCmd{}
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Prints draft supported values in machine-readable format",
		Long:  `This command prints information about the current draft environment and supported values such as supported dockerfile languages and deployment manifest types.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ic.run(); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&ic.format, "format", "f", ".", "specify the format to print draft information in (json, yaml, etc)")

	return cmd
}

func (ic *infoCmd) run() error {
	log.Debugf("getting supported languages")
	l := languages.CreateLanguagesFromEmbedFS(template.Dockerfiles, "")
	d := deployments.CreateDeploymentsFromEmbedFS(template.Deployments, "")

	ic.info = &draftInfo{
		SupportedLanguages:       l.Names(),
		SupportedDeploymentTypes: d.DeployTypes(),
	}

	infoText, err := json.MarshalIndent(ic.info, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal draft info into json: %w", err)
	}
	log.Println(string(infoText))
	return nil
}

func init() {
	rootCmd.AddCommand(newInfoCmd())
}
