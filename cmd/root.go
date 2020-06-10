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
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".servicemeow")
	}
	viper.SetEnvPrefix("sm")
	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
}
