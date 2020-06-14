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
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editChangeCmd represents the editChange command
var editChangeCmd = &cobra.Command{
	Use:   "change [change number]",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: editChange,
}

func init() {
	editCmd.AddCommand(editChangeCmd)
	editChangeCmd.Flags().StringP("output", "o", "report", "created change output type")
	editChangeCmd.Flags().StringP("file", "f", "", "input file")
	editChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")

	editChangeCmd.Flags().StringSlice("required", make([]string, 0), "required fields, comma seperated")
	viper.BindPFlag("edit_change_required", editChangeCmd.LocalFlags().Lookup("required"))
}

func editChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	viper.BindPFlag("file", cmd.Flags().Lookup("file"))
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))

	changeNumber := args[0]

	baseURL, _ := url.Parse(viper.GetString("servicenow.url"))

	serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: servicenow.DefaultEndpoints,
	}

	paramsMap := make(map[string]string, 0)
	var requestKoanf = koanf.New(".")
	requestKoanf.Load(file.Provider(viper.GetString("file")), yaml.Parser())
	requestGab := gabs.Wrap(requestKoanf.Raw())

	jsonMap := util.UntidyString(*requestGab)
	assignmentGroup := jsonMap.Search("assignment_group")
	if assignmentGroup != nil {
		assignmentGroupResp, err := findGroup(assignmentGroup.Data().(string))
		assignmentGroupRespGab, err := gabs.ParseJSON(assignmentGroupResp)

		if err != nil {
			//handle err
		}

		//sanity check result as ServiceNow may return all results if something doesn't match(!?)
		assignmentGroupNameGab, err := assignmentGroupRespGab.JSONPointer("/result/0/name")
		if assignmentGroupNameGab.Data().(string) != assignmentGroup.Data().(string) {
			fmt.Printf("Assignment group: \"%s\" not found", assignmentGroup.Data().(string))
			os.Exit(1)

		}
		assignmentGroupGab, err := assignmentGroupRespGab.JSONPointer("/result/0/sys_id")
		jsonMap.Set(assignmentGroupGab.Data(), "assignment_group")
	}

	requiredFieldErr := validateRequiredFields(jsonMap)
	if requiredFieldErr != nil {
		fmt.Println(requiredFieldErr)
		os.Exit(1)

	}

	paramsMap = make(map[string]string, 0)
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

	sysIDPath := path.Join(serviceNow.Endpoints["changeEndpoint"].Path, viper.GetString("type"), sysIDString)
	resp, err = serviceNow.HTTPRequest(serviceNow.Endpoints["changeEndpoint"], "PATCH", sysIDPath, nil, jsonMap.String())
	if err != nil {
		fmt.Println(err)
	}

	gabContainer, err = gabs.ParseJSON(resp)

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))

	}
}
