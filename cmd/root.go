/*
Copyright © 2020 DANIEL HOUSTON <houston@wehaveaproblem.co.uk>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var nocolor *bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "servicemeow",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if *nocolor {
			color.NoColor = true
		}
	},
	Short: "servicemeow - an unoffical ServiceNow cli",
	Long: `servicemeow is a cli for simplifying interacting with ServiceNow.
It handles both the creation, updating and processing of ServiceNow records
with configuration options suitable for automation. meow.`,
	// Errors are reported once, by Execute, so they are not duplicated and
	// usage text isn't dumped for runtime (non-usage) failures.
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command. Errors go to stderr so stdout stays clean
// for machine-readable output such as `-o raw`.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.servicemeow.yaml)")
	nocolor = rootCmd.PersistentFlags().Bool("nocolor", false, "disable color output")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			exitWithError(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".servicemeow")
	}
	viper.SetEnvPrefix("sm")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // servicenow.url -> SM_SERVICENOW_URL
	viper.AutomaticEnv()                                   // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		// The default config file is optional (settings may come from SM_*
		// env vars), but an explicit --config that can't be read, or a config
		// file that exists but is invalid, is fatal.
		var notFound viper.ConfigFileNotFoundError
		if cfgFile == "" && errors.As(err, &notFound) {
			return
		}
		exitWithError(fmt.Errorf("reading config: %w", err))
	}
}
