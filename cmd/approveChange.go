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

// approveChangeCmd represents the approveChange command
var approveChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: approveChange,
}

func init() {
	approveCmd.AddCommand(approveChangeCmd)
	approveChangeCmd.Flags().StringP("output", "o", "report", "change output type")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// approveChangeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// approveChangeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func approveChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	changeNumber := args[0]

	var changeEndpoint = &servicenow.Endpoint{
		Base:    "sn_chg_rest",
		Version: "v1",
		Path:    "change",
	}
	var tableEndpoint = &servicenow.Endpoint{
		Base:    "now",
		Version: "v1",
		Path:    "table/change_request",
	}

	endpoints := make(map[string]servicenow.Endpoint, 0)
	endpoints["changeEndpoint"] = *changeEndpoint
	endpoints["tableEndpoint"] = *tableEndpoint

	baseURL, _ := url.Parse(viper.GetString("servicenow.url"))

	serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: endpoints,
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

	changeEndpoint.Path = path.Join(serviceNow.Endpoints["changeEndpoint"].Path, sysIDString, "approvals")
	postResp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["changeEndpoint"], "PATCH", changeEndpoint.Path, nil, "{\"state\": \"approved\"}")
	gabContainer, err = gabs.ParseJSON(postResp)

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))

	}
}
