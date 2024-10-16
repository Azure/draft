package cmd

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/handlers"
)

type Format string

const (
	JSON Format = "json"
)

var supportedDeploymentTypes = [...]string{
	"helm",
	"kustomize",
	"manifests",
}

type infoCmd struct {
	format string
	info   *draftInfo
}

// draftConfigInfo is a struct that contains information about the example usage of variables for a single draft.yaml
type draftConfigInfo struct {
	Name                  string              `json:"name"`
	DisplayName           string              `json:"displayName,omitempty"`
	VariableExampleValues map[string][]string `json:"variableExampleValues,omitempty"`
}

type draftInfo struct {
	SupportedLanguages       []draftConfigInfo `json:"supportedLanguages"`
	SupportedDeploymentTypes [3]string         `json:"supportedDeploymentTypes"`
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
	supportedDockerfileTemplates := handlers.GetTemplatesByType(handlers.TemplateTypeDockerfile)

	languagesInfo := make([]draftConfigInfo, 0)
	for _, template := range supportedDockerfileTemplates {
		newConfig := draftConfigInfo{
			Name:                  template.Config.TemplateName,
			DisplayName:           template.Config.DisplayName,
			VariableExampleValues: template.Config.GetVariableExampleValues(),
		}
		languagesInfo = append(languagesInfo, newConfig)
	}

	ic.info = &draftInfo{
		SupportedLanguages:       languagesInfo,
		SupportedDeploymentTypes: supportedDeploymentTypes,
	}

	infoText, err := json.MarshalIndent(ic.info, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal draft info into json: %w", err)
	}
	fmt.Println(string(infoText))
	return nil
}

func init() {
	rootCmd.AddCommand(newInfoCmd())
}
