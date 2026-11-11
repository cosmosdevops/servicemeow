package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/CosmosDevops/servicemeow/servicenow"
	"github.com/CosmosDevops/servicemeow/util"
	"github.com/Jeffail/gabs/v2"
	"github.com/spf13/viper"
	"github.com/tj/go-naturaldate"
)

const serviceNowTimeLayout = "2006-01-02 15:04:05"

var validOutputTypes = []string{"report", "prettyjson", "raw"}

// newServiceNow builds a ServiceNow client from the configured URL.
func newServiceNow() (servicenow.ServiceNow, error) {
	rawURL := viper.GetString("servicenow.url")
	if rawURL == "" {
		return servicenow.ServiceNow{}, errors.New("servicenow.url is not set: add it to the config file or set SM_SERVICENOW_URL")
	}
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return servicenow.ServiceNow{}, fmt.Errorf("invalid servicenow.url: %w", err)
	}
	return servicenow.ServiceNow{
		BaseURL:   *baseURL,
		Endpoints: servicenow.DefaultEndpoints,
	}, nil
}

// validateOutput checks the --output flag before any request is made, so a
// typo doesn't result in a change being modified with nothing printed.
func validateOutput() error {
	output := viper.GetString("output")
	if !slices.Contains(validOutputTypes, output) {
		return fmt.Errorf("invalid output type %q, expected one of %v", output, validOutputTypes)
	}
	return nil
}

// changeRecord is a change request as returned by the Table API.
type changeRecord struct {
	SysID  string
	Type   string
	Record *gabs.Container
	Raw    []byte
}

// findChange looks up a change request by number, verifying that ServiceNow
// returned the requested change and not an unrelated record.
func findChange(sn servicenow.ServiceNow, number string) (*changeRecord, error) {
	endpoint := sn.Endpoints["tableEndpoint"]
	resp, err := sn.HTTPRequest(endpoint, "GET", endpoint.Path, map[string]string{"sysparm_query": "number=" + number}, "")
	if err != nil {
		return nil, err
	}
	parsed, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	record := parsed.S("result", "0")
	if record == nil {
		return nil, fmt.Errorf("change %q not found", number)
	}
	// ServiceNow may return unrelated records if the query doesn't match.
	if got, _ := record.S("number").Data().(string); !strings.EqualFold(got, number) {
		return nil, fmt.Errorf("change %q not found", number)
	}

	sysID, err := stringField(record, "sys_id")
	if err != nil {
		return nil, err
	}
	changeType, err := stringField(record, "type")
	if err != nil {
		return nil, err
	}
	return &changeRecord{SysID: sysID, Type: changeType, Record: record, Raw: resp}, nil
}

// stringField returns a non-empty string field from a record.
func stringField(record *gabs.Container, field string) (string, error) {
	if record == nil {
		return "", fmt.Errorf("unexpected ServiceNow response: missing %q", field)
	}
	value, ok := record.S(field).Data().(string)
	if !ok || value == "" {
		return "", fmt.Errorf("unexpected ServiceNow response: missing %q", field)
	}
	return value, nil
}

func parseResponse(resp []byte) (*gabs.Container, error) {
	parsed, err := gabs.ParseJSON(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to parse ServiceNow response: %w", err)
	}
	return parsed, nil
}

// writeResult prints a ServiceNow response in the configured output format.
// For report and prettyjson, hierarchy selects the part of the response to show.
func writeResult(resp []byte, hierarchy ...string) error {
	output := viper.GetString("output")
	if output == "raw" {
		fmt.Println(string(resp))
		return nil
	}
	parsed, err := parseResponse(resp)
	if err != nil {
		return err
	}
	result := parsed.S(hierarchy...)
	if result == nil {
		return fmt.Errorf("unexpected ServiceNow response: missing %q", strings.Join(hierarchy, "."))
	}
	util.WriteFormattedOutput(output, *result)
	return nil
}

// jsonBody encodes a request body, escaping user-supplied values correctly.
func jsonBody(fields map[string]string) string {
	body, err := json.Marshal(fields)
	if err != nil {
		// map[string]string always marshals
		panic(err)
	}
	return string(body)
}

// parseTime accepts either an absolute "YYYY-MM-DD HH:MM:SS" time or a
// natural language time relative to now, such as "tomorrow at 1pm".
func parseTime(flag, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("--%s must not be empty", flag)
	}
	if t, err := time.ParseInLocation(serviceNowTimeLayout, value, time.Local); err == nil {
		return t, nil
	}
	t, err := naturaldate.Parse(value, time.Now(), naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse --%s %q: use YYYY-MM-DD HH:MM:SS or a relative time such as \"tomorrow at 1pm\"", flag, value)
	}
	return t, nil
}

// parseTimeWindow parses the --start and --end flags.
func parseTimeWindow() (start, end time.Time, err error) {
	if start, err = parseTime("start", viper.GetString("start")); err != nil {
		return
	}
	if end, err = parseTime("end", viper.GetString("end")); err != nil {
		return
	}
	if end.Before(start) {
		err = fmt.Errorf("--end (%s) is before --start (%s)", end.Format(serviceNowTimeLayout), start.Format(serviceNowTimeLayout))
	}
	return
}

// resolveAssignmentGroup replaces an assignment_group name in a request with
// the group's sys_id, as ServiceNow requires.
func resolveAssignmentGroup(request *gabs.Container) error {
	group := request.Search("assignment_group")
	if group == nil {
		return nil
	}
	name, ok := group.Data().(string)
	if !ok {
		return fmt.Errorf("assignment group must be a group name, got %v", group.Data())
	}
	resp, err := findGroup(name)
	if err != nil {
		return err
	}
	parsed, err := parseResponse(resp)
	if err != nil {
		return err
	}
	sysID, err := stringField(parsed.S("result", "0"), "sys_id")
	if err != nil {
		return err
	}
	_, err = request.Set(sysID, "assignment_group")
	return err
}
