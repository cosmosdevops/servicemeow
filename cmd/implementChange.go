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
	"time"

	"github.com/DanielHouston/servicemeow/servicenow"
	"github.com/DanielHouston/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-naturaldate"
)

// implementChangeCmd represents the implementChange command
var implementChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: implementChange,
}

func init() {
	implementCmd.AddCommand(implementChangeCmd)
	implementChangeCmd.Flags().StringP("output", "o", "report", "created change output type")
	implementChangeCmd.Flags().StringP("start", "s", "now", "change type")
	implementChangeCmd.Flags().StringP("end", "e", "", "created change output type")
}
func implementChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("start", cmd.Flags().Lookup("start"))
	viper.BindPFlag("end", cmd.Flags().Lookup("end"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	changeNumber := args[0]

	var starttime time.Time
	starttime, err := naturaldate.Parse(viper.GetString("start"), time.Now(), naturaldate.WithDirection(naturaldate.Future))
	if err != nil {

		// error in parsing the date relatively, so pass it through directly
		starttime, err = time.Parse("2006-01-02 15:04:05", viper.GetString("start"))
	}

	var endtime time.Time
	endtime, err = naturaldate.Parse(viper.GetString("end"), time.Now(), naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		// error in parsing the date relatively, so pass it through directly
		endtime, err = time.Parse("2006-01-02 15:04:05", viper.GetString("end"))
	}

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

	changeType, err := gabContainer.JSONPointer("/result/0/type")
	viper.Set("type", changeType.String()[1:len(changeType.String())-1])

	tableEndpoint.Path = path.Join(tableEndpoint.Path, sysIDString)
	postBody := fmt.Sprintf("{\"work_start\": \"%s\",\n\"work_end\":\"%s\"}", starttime.Format("2006-01-02 15:04:05"), endtime.Format("2006-01-02 15:04:05"))
	postResp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["tableEndpoint"], "PATCH", tableEndpoint.Path, nil, postBody)

	gabContainer, err = gabs.ParseJSON(postResp)

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))

	}

}
