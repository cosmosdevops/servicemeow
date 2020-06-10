package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/camelcase"
)

func CamelCaseMap(inputMap gabs.Container) gabs.Container {
	newContainer := gabs.New()
	for key, child := range inputMap.ChildrenMap() {
		splitNameString := toCamelCase(key)
		if len(child.ChildrenMap()) != 0 {
			value := CamelCaseMap(*child)
			newContainer.ArrayAppend(value.Data(), splitNameString)
		} else {
			newContainer.Set(child.Data(), splitNameString)
		}
	}
	return *newContainer
}

var snakeCaseRegex = regexp.MustCompile("(^[A-Za-z]|_[A-Za-z])")

func toCamelCase(input string) string {
	return snakeCaseRegex.ReplaceAllStringFunc(input, func(s string) string {
		r := strings.NewReplacer("_", "")
		return strings.ToUpper(r.Replace(s))
	})
}

// unravel json into ServiceNow req friendly format
func UntidyString(inputMap gabs.Container) gabs.Container {
	untidyGab := gabs.New()

	for k, v := range inputMap.ChildrenMap() {

		camelCase := camelcase.Split(k)

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

		toCheck := strings.ToLower(splitNameString.String()) // lowercase to be compliant with SN requests
		if len(v.ChildrenMap()) != 0 {
			nestedObjectValue := UntidyString(*v)
			untidyGab.ArrayConcat(nestedObjectValue.Data(), toCheck)
		} else {
			untidyGab.Set(v.Data(), toCheck)
		}
	}
	fmt.Println(untidyGab)
	return *untidyGab
}
