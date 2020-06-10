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
	"path"
	"strings"

	"github.com/DanielHouston/servicemeow/servicenow"
	"github.com/DanielHouston/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/camelcase"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// addChangeCmd represents the addChange command
var addChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Add a change request",
	Long:  `Add a normal or standard change request`,
	Run:   addChange,
}

func init() {
	addCmd.AddCommand(addChangeCmd)

	addChangeCmd.Flags().StringP("type", "t", "normal", "change type")
	addChangeCmd.Flags().StringP("output", "o", "report", "created change output type")
	addChangeCmd.Flags().StringP("file", "f", "", "input file (required)")
	addChangeCmd.MarkFlagRequired("file")
	addChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")
	addChangeCmd.Flags().StringSlice("required", make([]string, 0), "required fields, comma seperated")

}

func addChange(cmd *cobra.Command, args []string) {
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))
	viper.BindPFlag("type", cmd.Flags().Lookup("type"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	viper.BindPFlag("required", cmd.Flags().Lookup("required"))
	viper.BindPFlag("file", cmd.Flags().Lookup("file"))

	path := path.Join("change", viper.GetString("type"))
	var changeEndpoint = &servicenow.Endpoint{
		Base:    "sn_chg_rest",
		Version: "v1",
		Path:    path,
	}

	endpoints := make(map[string]servicenow.Endpoint, 0)
	endpoints["changeEndpoint"] = *changeEndpoint

	baseURL, _ := url.Parse(viper.GetString("servicenow.url"))

	serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: endpoints,
	}
	paramsMap := make(map[string]string, 0)
	var requestKoanf = koanf.New(".")

	requestKoanf.Load(file.Provider(viper.GetString("file")), yaml.Parser())
	requestGab := gabs.Wrap(requestKoanf.Raw())

	jsonMap := util.UntidyString(*requestGab)
	assignmentGroup := jsonMap.Search("assignment_group")
	if assignmentGroup != nil {
		assignmentGroupResp := findGroup(assignmentGroup.Data().(string))
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

	fmt.Print(jsonMap)
	requiredFieldErr := validateRequiredFields(jsonMap)
	if requiredFieldErr != nil {
		fmt.Println(requiredFieldErr)
		os.Exit(1)

	}
	resp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["changeEndpoint"], "POST", path, paramsMap, jsonMap.String())
	if err != nil {
		fmt.Println(err)
	}
	gabContainer, err := gabs.ParseJSON(resp)

	if viper.GetString("output") == "raw" {
		fmt.Println(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result"))

	}

}

func validateRequiredFields(reqToCheck gabs.Container) error {
	fmt.Print(reqToCheck)
	splitRequiredFields := make([]string, 0)
	for _, eachRequiredField := range viper.GetStringSlice("required") {
		splitRequiredFields = append(splitRequiredFields, strings.Split(eachRequiredField, ",")...)
	}
	for requiredFieldIndex := 0; requiredFieldIndex < len(splitRequiredFields); requiredFieldIndex++ {
		fieldToCheck := splitRequiredFields[requiredFieldIndex]
		camelCase := camelcase.Split(fieldToCheck)

		var splitNameString strings.Builder
		for i, eachSplit := range camelCase {
			switch eachSplit {
			case "_":
			case " ":
			case "-":
			default:
				splitNameString.WriteString(eachSplit)
				if i != len(camelCase)-1 {
					splitNameString.WriteString("_")
				}
			}
		}

		toCheck := strings.ToLower(splitNameString.String())
		if !reqToCheck.Exists(toCheck) {
			return fmt.Errorf("ERROR: Change definition missing required field \"%s\"", fieldToCheck)
		}

	}
	return nil
}
