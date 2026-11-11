/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cancelChangeCmd represents the `cancel change` command
var cancelChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Cancel a change request",
	Long:  `Cancel a change by moving it into the Cancelled state`,
	RunE:  cancelChange,
}

func init() {
	cancelCmd.AddCommand(cancelChangeCmd)
	cancelChangeCmd.Flags().StringP("reason", "r", "", "reason why the change has been canceled (required)")
	cancelChangeCmd.MarkFlagRequired("reason")
	cancelChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
}

func cancelChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("reason", cmd.Flags().Lookup("reason"))
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

	endpoint := sn.Endpoints["changeEndpoint"]
	changePath := path.Join(endpoint.Path, change.Type, change.SysID)
	body := jsonBody(map[string]string{"state": "Canceled", "work_notes": viper.GetString("reason")})
	resp, err := sn.HTTPRequest(endpoint, "PATCH", changePath, nil, body)
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
