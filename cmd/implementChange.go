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
	"time"

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/openshift/origin/pkg/cmd/server/start"
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
	RunE: implementChange,
}

func init() {
	implementCmd.AddCommand(implementChangeCmd)
	implementChangeCmd.Flags().StringP("output", "o", "report", "created change output type")
	implementChangeCmd.Flags().StringP("start", "s", "now", "change type")
	implementChangeCmd.Flags().StringP("end", "e", "", "created change output type")
}
func implementChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("start", cmd.Flags().Lookup("start"))
	viper.BindPFlag("end", cmd.Flags().Lookup("end"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	changeNumber := args[0]

	var starttime time.Time
	starttime, err := naturaldate.Parse(viper.GetString("start"), time.Now(), naturaldate.WithDirection(naturaldate.Future))
	if viper.GetString("start") != "now" && starttime.Equal(time.Now()){
		return errors.New("Invalid start time")
	}

	if err != nil {
		// error in parsing the date relatively, so pass it through directly
		starttime, err = time.Parse("2006-01-02 15:04:05", viper.GetString("start"))
		if err != nil {
			return err
		}

	}

	var endtime time.Time
	endtime, err = naturaldate.Parse(viper.GetString("end"), time.Now(), naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		// error in parsing the date relatively, so pass it through directly
		endtime, err = time.Parse("2006-01-02 15:04:05", viper.GetString("end"))
		if err != nil {
			return err
		}
	}

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
	if err != nil {
		return err
	}

	changeType, err := gabContainer.JSONPointer("/result/0/type")
	if err != nil {
		return err
	}
	viper.Set("type", changeType.String()[1:len(changeType.String())-1])

	sysIDPath := path.Join(serviceNow.Endpoints["tableEndpoint"].Path, sysIDString)
	postBody := fmt.Sprintf("{\"work_start\": \"%s\",\n\"work_end\":\"%s\"}", starttime.Format("2006-01-02 15:04:05"), endtime.Format("2006-01-02 15:04:05"))
	postResp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["tableEndpoint"], "PATCH", sysIDPath, nil, postBody)
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
