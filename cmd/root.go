package cmd

import (
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Azure/draft/pkg/logger"
)

var cfgFile string
var verbose bool
var provider string
var silent bool
var dryRun bool
var dryRunFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "draft",
	Short: "Draft creates the minimum required files for your application to run on Kubernetes",
	Long: `Draft is a Command Line Tool (CLI) that creates the miminum required files for your Kubernetes deployments.

To start a k8s deployment with draft, run the 'draft create' command 🤩

	$ draft create

This will prompt you to create a Dockerfile and deployment files for your project ✨

For more information, please visit the Draft Github page: https://github.com/Azure/draft.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		} else if silent {
			logrus.SetLevel(logrus.ErrorLevel)

		}
		logrus.SetOutput(&logger.OutputSplitter{})
		logrus.SetFormatter(new(logger.CustomFormatter))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cc.Init(&cc.Config{
		RootCmd:  rootCmd,
		Headings: cc.Cyan + cc.Bold + cc.Underline,
		Commands: cc.Bold,
		Example:  cc.Italic,
		ExecName: cc.Bold,
		Flags:    cc.Bold,
	})
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.draft.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "azure", "cloud provider")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "", false, "enable silent logging")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "enable dry run mode in which no files are written to disk")
	rootCmd.PersistentFlags().StringVar(&dryRunFile, "dry-run-file", "", "optional file to write dry run summary in json format into (requires --dry-run flag)")
}
