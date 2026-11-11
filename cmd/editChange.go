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

// editChangeCmd represents the editChange command
var editChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Edit a change request",
	Long:  `Edit a change request by patching an existing change with values taken from an input file.`,
	RunE:  editChange,
}

func init() {
	editCmd.AddCommand(editChangeCmd)
	editChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
	editChangeCmd.Flags().StringP("file", "f", "", "input file (required)")
	editChangeCmd.MarkFlagRequired("file")
	editChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")

	editChangeCmd.Flags().StringSlice("required", make([]string, 0), "required fields, comma seperated")
}

func editChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	viper.BindPFlag("file", cmd.Flags().Lookup("file"))
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))
	viper.BindPFlag("required", cmd.Flags().Lookup("required"))
	if err := validateOutput(); err != nil {
		return err
	}

	sn, err := newServiceNow()
	if err != nil {
		return err
	}

	request, err := loadRequestFile(viper.GetString("file"))
	if err != nil {
		return err
	}
	if err := resolveAssignmentGroup(request); err != nil {
		return err
	}
	if err := validateRequiredFields(*request); err != nil {
		return err
	}

	change, err := findChange(sn, args[0])
	if err != nil {
		return err
	}

	endpoint := sn.Endpoints["changeEndpoint"]
	resp, err := sn.HTTPRequest(endpoint, "PATCH", path.Join(endpoint.Path, change.Type, change.SysID), nil, request.String())
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
