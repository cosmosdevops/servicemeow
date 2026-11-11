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

// rejectChangeCmd represents the `reject change` command
var rejectChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Reject a change request",
	Long: `Reject a change request and move it to the close state.
A comment is required to explain why the change was rejected.
Change will be rejected as the current user.`,
	RunE: rejectChange,
}

func init() {
	rejectCmd.AddCommand(rejectChangeCmd)
	rejectChangeCmd.Flags().StringP("comment", "c", "", "comment describing why the change is rejected (required)")
	rejectChangeCmd.MarkFlagRequired("comment")
	rejectChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
}

func rejectChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("comment", cmd.Flags().Lookup("comment"))
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
	approvalsPath := path.Join(endpoint.Path, change.SysID, "approvals")
	body := jsonBody(map[string]string{"state": "rejected", "comments": viper.GetString("comment")})
	resp, err := sn.HTTPRequest(endpoint, "PATCH", approvalsPath, nil, body)
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
