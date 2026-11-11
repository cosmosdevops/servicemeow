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
	"path"
	"strings"

	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/camelcase"
	"github.com/knadh/koanf/v2"
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
	RunE:  addChange,
}

func init() {
	addCmd.AddCommand(addChangeCmd)

	addChangeCmd.Flags().StringP("type", "t", "normal", "change type")
	addChangeCmd.Flags().StringP("output", "o", "report", "output type: report, prettyjson or raw")
	addChangeCmd.Flags().StringP("file", "f", "", "input file (required)")
	addChangeCmd.MarkFlagRequired("file")
	addChangeCmd.Flags().Bool("showempty", false, "show all fields even if they are empty")
	addChangeCmd.Flags().StringSlice("required", make([]string, 0), "required fields, comma seperated")

}

func addChange(cmd *cobra.Command, args []string) error {
	viper.BindPFlag("showempty", cmd.Flags().Lookup("showempty"))
	viper.BindPFlag("type", cmd.Flags().Lookup("type"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	viper.BindPFlag("required", cmd.Flags().Lookup("required"))
	viper.BindPFlag("file", cmd.Flags().Lookup("file"))
	if err := validateOutput(); err != nil {
		return err
	}

	sn, err := newServiceNow()
	if err != nil {
		return err
	}

	request, err := loadRequestFile(viper.GetString("file"))
	if err != nil {
		return err
	}
	if err := resolveAssignmentGroup(request); err != nil {
		return err
	}
	if err := validateRequiredFields(*request); err != nil {
		return err
	}

	endpoint := sn.Endpoints["changeEndpoint"]
	resp, err := sn.HTTPRequest(endpoint, "POST", path.Join(endpoint.Path, viper.GetString("type")), nil, request.String())
	if err != nil {
		return err
	}
	return writeResult(resp, "result")
}

// loadRequestFile reads a YAML request definition and converts its keys to
// the snake_case field names ServiceNow expects.
func loadRequestFile(filename string) (*gabs.Container, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(filename), yaml.Parser()); err != nil {
		return nil, err
	}
	request := util.UntidyString(*gabs.Wrap(k.Raw()))
	return &request, nil
}

func validateRequiredFields(reqToCheck gabs.Container) error {
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
			return fmt.Errorf("Definition missing required field \"%s\"", fieldToCheck)
		}

	}
	return nil
}
