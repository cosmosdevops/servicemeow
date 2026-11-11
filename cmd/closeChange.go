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

// closeChangeCmd represents the `close change` command
var closeChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Close a change request",
	Long:  `Close a change request by moving it to the Closed state. Requires a close code to indicate the reason for closing`,
	RunE:  closeChange,
}

func init() {
	closeCmd.AddCommand(closeChangeCmd)
	closeChangeCmd.Flags().StringP("note", "n", "", "description of the state of the change at closure (required)")
	closeChangeCmd.MarkFlagRequired("note")
	closeChangeCmd.Flags().StringP("code", "c", "", "closure code of the change (required)")
	closeChangeCmd.MarkFlagRequired("code")
	closeChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
}

func closeChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("note", cmd.Flags().Lookup("note"))
	viper.BindPFlag("code", cmd.Flags().Lookup("code"))
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
	body := jsonBody(map[string]string{
		"close_notes": viper.GetString("note"),
		"close_code":  viper.GetString("code"),
		"state":       "closed",
	})
	resp, err := sn.HTTPRequest(endpoint, "PATCH", changePath, nil, body)
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
