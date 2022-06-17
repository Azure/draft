package cmd

import (
	"fmt"
	"os"

	"github.com/Azure/draft/pkg/logger"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var silent bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "draft",
	Short: "Draft creates the minimum required files for your application to run on Kubernetes",
	Long: `Draft is a Command Line Tool (CLI) that creates the miminum required files for your Kubernetes deployments.

To start a k8s deployment with draft, run the 'draft create' command 🤩

	$ draft create

This will prompt you to create a Dockerfile and deployment files for your project ✨

For more information, please visit the Draft GitHub repo: https://github.com/Azure/draft.`,

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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.draft.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "", false, "enable silent logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".draft" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".draft")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
