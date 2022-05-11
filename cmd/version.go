package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"runtime/debug"
)

var VERSION = "v0.0.7"

func newVersionCmd() *cobra.Command {
	// versionCmd represents the version command
	var version = &cobra.Command{
		Use:   "version",
		Short: "Get the current version of Draft",
		Long:  `Returns the running version of Draft`,
		RunE: func(cmd *cobra.Command, args []string) error {

			vcsInfo := getVCSInfoFromRuntime()

			fmt.Println("version: ", VERSION)
			fmt.Println("runtime SHA: ", vcsInfo)
			return nil
		},
	}

	return version

}

func getVCSInfoFromRuntime() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatal("could not get vcs info at runtime")
	}

	for _, kv := range buildInfo.Settings {
		if kv.Key == "vcs.revision" {
			return kv.Value
		}
	}

	return ""
}

func init() {
	rootCmd.AddCommand(newVersionCmd())
}
