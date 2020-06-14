package servicenow

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/spf13/viper"
)

type ServiceNow struct {
	BaseURL   url.URL             `json:"url"`
	Endpoints map[string]Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Base    string `json:name`
	Version string `json:version`
	Path    string `json:path`
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, nil
}
