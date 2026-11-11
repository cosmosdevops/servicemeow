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
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getChangeCmd represents the `get change` command
var getChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Get a change request",
	Long: `Gets information on a change request based on its Change number, typically of format: CHGxxxxxxx.
The output will be formated for reading in the terminal by default. Using '-o raw' will provide the result as JSON.
For example:
 servicemeow get change -o raw CHG0000001`,
	RunE: getChange,
}

func init() {
	getCmd.AddCommand(getChangeCmd)
	getChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
	getChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")
}

func getChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	if err := validateOutput(); err != nil {
		return err
	}

	sn, err := newServiceNow()
	if err != nil {
		return err
	}
	change, err := findChange(sn, args[0])
	if err != nil {
		return err
	}

	if viper.GetString("output") == "raw" {
		return writeResult(change.Raw)
	}
	util.WriteFormattedOutput(viper.GetString("output"), *change.Record)
	return nil
}
