package servicenow

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) ServiceNow {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL + "/api")
	if err != nil {
		t.Fatal(err)
	}
	return ServiceNow{BaseURL: *u, Endpoints: DefaultEndpoints}
}

func TestHTTPRequestBuildsURLAndAuth(t *testing.T) {
	viper.Set("servicenow.username", "user")
	viper.Set("servicenow.password", "pass")
	t.Cleanup(viper.Reset)

	sn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/now/v1/table/change_request" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("sysparm_query"); got != "number=CHG1" {
			t.Errorf("sysparm_query = %q", got)
		}
		if u, p, ok := r.BasicAuth(); !ok || u != "user" || p != "pass" {
			t.Errorf("basic auth = %q %q %v", u, p, ok)
		}
		w.Write([]byte(`{"result":[]}`))
	})

	ep := sn.Endpoints["tableEndpoint"]
	body, err := sn.HTTPRequest(ep, "GET", ep.Path, map[string]string{"sysparm_query": "number=CHG1"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"result":[]}` {
		t.Errorf("body = %q", body)
	}
}

func TestHTTPRequestAcceptsAny2xx(t *testing.T) {
	for _, code := range []int{200, 201, 204} {
		sn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		})
		ep := sn.Endpoints["changeEndpoint"]
		if _, err := sn.HTTPRequest(ep, "POST", ep.Path, nil, "{}"); err != nil {
			t.Errorf("status %d: unexpected error %v", code, err)
		}
	}
}

func TestHTTPRequestErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"servicenow error", 404, `{"error":{"message":"No Record found","detail":"Record doesn't exist"},"status":"failure"}`, "ServiceNow returned 404 Not Found: No Record found: Record doesn't exist"},
		{"plain body", 401, "Unauthorized", "ServiceNow returned 401 Unauthorized: Unauthorized"},
		{"empty body", 500, "", "ServiceNow returned 500 Internal Server Error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			})
			ep := sn.Endpoints["tableEndpoint"]
			_, err := sn.HTTPRequest(ep, "GET", ep.Path, nil, "")
			if err == nil || err.Error() != tt.want {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestHTTPRequestTimeout(t *testing.T) {
	viper.Set("servicenow.timeout", "50ms")
	t.Cleanup(viper.Reset)

	done := make(chan struct{})
	sn := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	})
	defer close(done)

	ep := sn.Endpoints["tableEndpoint"]
	start := time.Now()
	_, err := sn.HTTPRequest(ep, "GET", ep.Path, nil, "")
	if err == nil || !strings.Contains(err.Error(), "Client.Timeout") {
		t.Fatalf("err = %v, want timeout", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("request took %v, timeout not applied", elapsed)
	}
}
