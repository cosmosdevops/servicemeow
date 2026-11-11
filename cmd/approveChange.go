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

// approveChangeCmd represents the `approve change` command
var approveChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Approve a change request",
	Long:  `Approve a change request which is either in the Assess or Approval stage, as the current user`,
	RunE:  approveChange,
}

func init() {
	approveCmd.AddCommand(approveChangeCmd)
	approveChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
}

func approveChange(cmd *cobra.Command, args []string) error {
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
	approvalPath := path.Join(endpoint.Path, change.SysID, "approvals")
	resp, err := sn.HTTPRequest(endpoint, "PATCH", approvalPath, nil, jsonBody(map[string]string{"state": "approved"}))
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
