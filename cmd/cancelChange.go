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
	"fmt"
	"net/url"
	"path"

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cancelChangeCmd represents the cancelChange command
var cancelChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "Cancel a change request",
	Long:  `Cancel a change by moving it into the Cancelled state`,
	RunE:   cancelChange,

}

func init() {
	cancelCmd.AddCommand(cancelChangeCmd)

	cancelChangeCmd.Flags().StringP("reason", "r", "", "reason why the change has been canceled (required)")
	cancelChangeCmd.MarkFlagRequired("reason")
	cancelChangeCmd.Flags().StringP("output", "o", "report", "change output type")
}

func cancelChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("reason", cmd.Flags().Lookup("reason"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	changeNumber := args[0]

	baseURL, err := url.Parse(viper.GetString("servicenow.url"))

	if err != nil {
		return err
	}

	serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: servicenow.DefaultEndpoints,
	}

	paramsMap := make(map[string]string, 0)
	paramsMap["sysparm_query"] = "number=" + changeNumber
	resp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["tableEndpoint"], "GET", serviceNow.Endpoints["tableEndpoint"].Path, paramsMap, "")
	if err != nil {
		return err
	}

	gabContainer, err := gabs.ParseJSON(resp)
	if err != nil {
		return err
	}

	sysID, err := gabContainer.JSONPointer("/result/0/sys_id")
	if err != nil {
		return err
	}

	sysIDString := sysID.String()[1 : len(sysID.String())-1]

	changeType, err := gabContainer.JSONPointer("/result/0/type")
	if err != nil {
		return err
	}

	changeTypeString := changeType.String()[1 : len(changeType.String())-1]

	sysIDPath := path.Join(serviceNow.Endpoints["changeEndpoint"].Path, changeTypeString, sysIDString)
	postResp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["changeEndpoint"], "PATCH", sysIDPath, nil, fmt.Sprintf("{\"state\":\"Canceled\",\"work_notes\":\"%s\"}", viper.GetString("reason")))
	if err != nil {
		return err
	}

	gabContainer, err = gabs.ParseJSON(postResp)
	if err != nil {
		return err
	}

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))
	}
	return nil
}
