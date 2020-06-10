package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/camelcase"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

func GenerateReport(data gabs.Container) string {
	bold := color.New(color.Bold).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	var str strings.Builder

	for key, child := range data.ChildrenMap() {
		keySplit := camelcase.Split(key)
		spacedKey := strings.Join(keySplit, " ")
		if len(child.ChildrenMap()) != 0 {
			value := GenerateReport(*child)
			if strings.Trim(value, "\"") == "" && viper.GetBool("showempty") == false {
				continue
			}
			str.WriteString(bold(spacedKey))
			str.WriteString("\n")
			str.WriteString(red(value))
			str.WriteString("\n")
		} else if len(child.Children()) > 0 {
			nested := child.Children()
			for i := 0; i < len(nested); i++ {
				value := GenerateReport(*nested[i])
				if strings.Trim(value, "\"") == "" && viper.GetBool("showempty") == false {
					continue
				}
				str.WriteString(bold(spacedKey))
				str.WriteString("\n")
				for _, line := range strings.Split(strings.TrimSuffix(value, "\n"), "\n") {
					str.WriteString("\t" + strings.Trim(line, "\""))
					str.WriteString("\n")
				}
				str.WriteString("\n")

			}

		} else {
			if strings.Trim(child.String(), "\"") == "" && viper.GetBool("showempty") == false {
				continue
			}
			str.WriteString(bold(spacedKey))
			str.WriteString("\n")
			str.WriteString(red(strings.Trim(child.String(), "\"")))
			str.WriteString("\n")
		}
	}
	return str.String()
}

func WriteFormattedOutput(outputType string, resp gabs.Container) {
	tidyMap := CamelCaseMap(resp)
	switch outputType {
	case "prettyjson":
		fmt.Println(tidyMap.StringIndent("", "  "))
	case "report":
		formattedMap := GenerateReport(tidyMap)
		fmt.Printf("%s", formattedMap)
	default:
	}
	os.Exit(0)
}
