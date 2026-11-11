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

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getGroupCmd represents the `get group` command
var getGroupCmd = &cobra.Command{
	Use:   "group [group name]",
	Args:  cobra.ExactArgs(1),
	Short: "Get a ServiceNow group",
	Long:  `Get a ServiceNow group by name`,
	RunE:  getGroup,
}

var userGroupTableEndpoint = servicenow.Endpoint{
	Base:    "now",
	Version: "v1",
	Path:    "table/sys_user_group",
}

func init() {
	getCmd.AddCommand(getGroupCmd)
	getGroupCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
}

func getGroup(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	if err := validateOutput(); err != nil {
		return err
	}

	resp, err := findGroup(args[0])
	if err != nil {
		return err
	}
	return writeResult(resp, "result", "0")
}

// findGroup looks up a user group by name, verifying that ServiceNow returned
// the requested group and not an unrelated record.
func findGroup(name string) ([]byte, error) {
	sn, err := newServiceNow()
	if err != nil {
		return nil, err
	}

	params := map[string]string{"sysparm_query": "name=" + name}
	resp, err := sn.HTTPRequest(userGroupTableEndpoint, "GET", userGroupTableEndpoint.Path, params, "")
	if err != nil {
		return nil, err
	}
	parsed, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// ServiceNow may return unrelated records if the query doesn't match.
	if got, _ := parsed.S("result", "0", "name").Data().(string); got != name {
		return nil, fmt.Errorf("assignment group %q not found", name)
	}
	return resp, nil
}
