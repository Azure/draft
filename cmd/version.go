package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"runtime/debug"
)

var VERSION = "v0.0.7"

func newVersionCmd() *cobra.Command {
	// versionCmd represents the version command
	var version = &cobra.Command{
		Use:   "version",
		Short: "Get current version of Draft",
		Long:  `Returns the running version of Draft`,
		RunE: func(cmd *cobra.Command, args []string) error {

			getVersionFromRuntime()

			log.Infof("version: %s", VERSION)
			return nil
		},
	}

	return version

}

func getVersionFromRuntime() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatal("could not get version at runtime")
	}
	for _, kv := range buildInfo.Settings {
		log.Infof("key: %s", kv.Key)
		log.Infof("value: %s", kv.Value)
	}
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
