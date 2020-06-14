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

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getGroupCmd represents the getGroup command
var getGroupCmd = &cobra.Command{
	Use:   "group [group name]",
	Args:  cobra.ExactArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: getGroup,
}

func init() {
	getCmd.AddCommand(getGroupCmd)
	getGroupCmd.Flags().StringP("output", "o", "report", "change output type")
}

func getGroup(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	groupName := args[0]

	resp, err := findGroup(groupName)

	if err != nil {
		return err
	}
	gabContainer, err := gabs.ParseJSON(resp)

	if err != nil {
		return err
	}

	if viper.GetString("output") == "raw" {
		fmt.Print(string(resp))
	} else {
		util.WriteFormattedOutput(viper.GetString("output"), *gabContainer.S("result", "0"))
	}
	return nil
}

func findGroup(name string) ([]byte, error) {

	var userGroupTableEndpoint = &servicenow.Endpoint{
		Base:    "now",
		Version: "v1",
		Path:    "table/sys_user_group",
	}

	endpoints := make(map[string]servicenow.Endpoint, 0)
	endpoints["userGroupTableEndpoint"] = *userGroupTableEndpoint

	baseURL, _ := url.Parse(viper.GetString("servicenow.url"))
	var serviceNow = servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: endpoints,
	}

	paramsMap := make(map[string]string, 0)
	paramsMap["sysparm_query"] = "name=" + name
	resp, err := serviceNow.HTTPRequest(serviceNow.Endpoints["userGroupTableEndpoint"], "GET", serviceNow.Endpoints["userGroupTableEndpoint"].Path, paramsMap, "")
	if err != nil {
		return nil, err
	}
	return resp, nil
}
