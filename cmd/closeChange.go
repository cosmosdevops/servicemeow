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
	"os"
	"path"

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// closeChangeCmd represents the closeChange command
var closeChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: closeChange,
}

func init() {
	closeCmd.AddCommand(closeChangeCmd)

	closeChangeCmd.Flags().StringP("note", "n", "", "description of the state of the change at closure (required)")
	closeChangeCmd.MarkFlagRequired("note")
	closeChangeCmd.Flags().StringP("code", "c", "", "closure code of the change (required)")
	closeChangeCmd.MarkFlagRequired("code")
	closeChangeCmd.Flags().StringP("output", "o", "report", "change output type")
}

func closeChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("note", cmd.Flags().Lookup("note"))
	viper.BindPFlag("code", cmd.Flags().Lookup("code"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	changeNumber := args[0]

	baseURL, _ := url.Parse(viper.GetString("servicenow.url"))

	serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: servicenow.DefaultEndpoints,
	}

	paramsMap := make(map[string]string, 0)
	paramsMap["sysparm_query"] = "number=" + changeNumber
	resp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["tableEndpoint"], "GET", serviceNow.Endpoints["tableEndpoint"].Path, paramsMap, "")
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	gabContainer, err := gabs.ParseJSON(resp)
	sysID, err := gabContainer.JSONPointer("/result/0/sys_id")
	sysIDString := sysID.String()[1 : len(sysID.String())-1]
	if err != nil {
		panic(err)
	}

	changeType, err := gabContainer.JSONPointer("/result/0/type")
	changeTypeString := changeType.String()[1 : len(changeType.String())-1]

	sysIDPath := path.Join(serviceNow.Endpoints["changeEndpoint"].Path, changeTypeString, sysIDString)
	postBody := fmt.Sprintf("{\"close_notes\": \"%s\",\n\"close_code\":\"%s\",\"state\": \"%s\"}", viper.GetString("note"), viper.GetString("code"), "closed")
	postResp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["changeEndpoint"], "PATCH", sysIDPath, nil, postBody)

	gabContainer, err = gabs.ParseJSON(postResp)

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))

	}
}
