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
	"fmt"
	"net/url"
	"os"

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/gommon/log"
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
 servicemeow get change --type -o raw standard CHG0000001`,
	Run: getChange,
}
var serviceNow servicenow.ServiceNow

func init() {
	getCmd.AddCommand(getChangeCmd)
	getChangeCmd.Flags().StringP("output", "o", "report", "change output type")
	getChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")

}

func getChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	var validOutputTypes []string = make([]string, 0)
	var valid bool = false
	validOutputTypes = append(validOutputTypes, "report", "prettyjson", "raw")
	for i := 0; i < len(validOutputTypes); i++ {
		if validOutputTypes[i] == viper.GetString("output") {
			valid = true
		}
	}
	if !valid {
		fmt.Printf("Invalid output type specified: %s. Try %v \n", viper.GetString("output"), validOutputTypes)
		os.Exit(1)
	}

	changeNumber := args[0]
	var tableEndpoint = &servicenow.Endpoint{
		Base:    "now",
		Version: "v1",
		Path:    "table/change_request",
	}

	endpoints := make(map[string]servicenow.Endpoint, 0)
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

	if err != nil {
		panic(err)
	}

	if viper.GetString("output") == "raw" {
		fmt.Print(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result", "0"))
	}
}
