package servicenow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// DefaultTimeout bounds each request so an unresponsive instance can't hang
// the CLI. Override with the servicenow.timeout config key (e.g. "2m").
const DefaultTimeout = 30 * time.Second

type ServiceNow struct {
	BaseURL   url.URL             `json:"url"`
	Endpoints map[string]Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Base    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

var changeEndpoint = &Endpoint{
	Base:    "sn_chg_rest",
	Version: "v1",
	Path:    "change",
}
var tableEndpoint = &Endpoint{
	Base:    "now",
	Version: "v1",
	Path:    "table/change_request",
}
var DefaultEndpoints = map[string]Endpoint{"tableEndpoint": *tableEndpoint, "changeEndpoint": *changeEndpoint}

func (s ServiceNow) buildURL(endpoint *Endpoint, urlPath *string, params map[string]string) *url.URL {
	requestURL := s.BaseURL

	requestURL.Path = path.Join(requestURL.Path, endpoint.Base)
	requestURL.Path = path.Join(requestURL.Path, endpoint.Version)
	requestURL.Path = path.Join(requestURL.Path, *urlPath)
	query := requestURL.Query()

	for key, value := range params {
		query.Add(key, value)
	}
	requestURL.RawQuery = query.Encode()

	return &requestURL
}

func (s ServiceNow) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(viper.GetString("servicenow.username"), viper.GetString("servicenow.password"))
}

func (s ServiceNow) HTTPRequest(endpoint Endpoint, method string, urlPath string, params map[string]string, reqBody string) ([]byte, error) {
	reqURL := s.buildURL(&endpoint, &urlPath, params)
	req, err := http.NewRequest(method, reqURL.String(), strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	s.setHeaders(req)

	timeout := viper.GetDuration("servicenow.timeout")
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ServiceNow response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, responseError(resp, body)
	}
	return body, nil
}

// responseError builds an error from a failed request, using ServiceNow's
// error message and detail when the body contains them.
func responseError(resp *http.Response, body []byte) error {
	var snErr struct {
		Error struct {
			Message string `json:"message"`
			Detail  string `json:"detail"`
		} `json:"error"`
	}
	msg := strings.TrimSpace(string(body))
	if json.Unmarshal(body, &snErr) == nil && snErr.Error.Message != "" {
		msg = snErr.Error.Message
		if snErr.Error.Detail != "" {
			msg += ": " + snErr.Error.Detail
		}
	}
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}
	if msg == "" {
		return fmt.Errorf("ServiceNow returned %s", resp.Status)
	}
	return fmt.Errorf("ServiceNow returned %s: %s", resp.Status, msg)
}
