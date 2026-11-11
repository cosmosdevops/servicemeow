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

// scheduleChangeCmd represents the scheduleChange command
var scheduleChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Schedule a change request",
	Long: `Schedule a change request between --start and --end time.
Times can be provided in a traditional date/time format of YYYY-MM-DD HH:MM:SS
or can be relative from the current time with such grammar as:
   now
   tomorrow at 1pm
   22nd of December
   yesterday
etc.
	
Language which is not understood is ignored. This can have unintented consequences with typos as:
  --start "22nd Decmber "
would resolve to "22nd" of the current month as "Decmber" would be ignored."`,
	RunE: scheduleChange,
}

func init() {
	scheduleCmd.AddCommand(scheduleChangeCmd)
	scheduleChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
	scheduleChangeCmd.Flags().StringP("start", "s", "now", "planned start time")
	scheduleChangeCmd.Flags().StringP("end", "e", "", "planned end time (required)")
	scheduleChangeCmd.MarkFlagRequired("end")
}

func scheduleChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("start", cmd.Flags().Lookup("start"))
	viper.BindPFlag("end", cmd.Flags().Lookup("end"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	if err := validateOutput(); err != nil {
		return err
	}

	start, end, err := parseTimeWindow()
	if err != nil {
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
	body := jsonBody(map[string]string{
		"start_date": start.Format(serviceNowTimeLayout),
		"end_date":   end.Format(serviceNowTimeLayout),
	})
	resp, err := sn.HTTPRequest(endpoint, "PATCH", path.Join(endpoint.Path, change.Type, change.SysID), nil, body)
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}
