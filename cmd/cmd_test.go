package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
}

// fakeServiceNow is a minimal stand-in for the ServiceNow REST APIs used by
// the CLI. It knows one change (CHG0000001) and one group (Ops).
type fakeServiceNow struct {
	t        *testing.T
	mu       sync.Mutex
	requests []recordedRequest
	// writeStatus, when set, is returned for PATCH/POST requests.
	writeStatus int
}

func (f *fakeServiceNow) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{r.Method, r.URL.Path, r.URL.Query().Get("sysparm_query"), string(body)})
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("sysparm_query")
	switch {
	case r.Method == "GET" && r.URL.Path == "/api/now/v1/table/change_request":
		switch query {
		case "number=CHG0000001":
			w.Write([]byte(`{"result":[{"number":"CHG0000001","sys_id":"abc123","type":"normal","short_description":"Patch servers"}]}`))
		case "number=CHGMISMATCH":
			w.Write([]byte(`{"result":[{"number":"CHG9999999","sys_id":"other","type":"normal"}]}`))
		default:
			w.Write([]byte(`{"result":[]}`))
		}
	case r.Method == "GET" && r.URL.Path == "/api/now/v1/table/sys_user_group":
		if query == "name=Ops" {
			w.Write([]byte(`{"result":[{"name":"Ops","sys_id":"grp1"}]}`))
		} else {
			w.Write([]byte(`{"result":[]}`))
		}
	case r.Method == "PATCH" || r.Method == "POST":
		if f.writeStatus != 0 {
			w.WriteHeader(f.writeStatus)
			w.Write([]byte(`{"error":{"message":"Operation failed","detail":"State transition not allowed"},"status":"failure"}`))
			return
		}
		w.Write([]byte(`{"result":{"number":"CHG0000001","state":"updated"}}`))
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeServiceNow) writes() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []recordedRequest
	for _, r := range f.requests {
		if r.Method != "GET" {
			out = append(out, r)
		}
	}
	return out
}

// run executes the CLI against a fake ServiceNow and returns stdout.
func run(t *testing.T, fake *fakeServiceNow, args ...string) (string, error) {
	t.Helper()
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfg, []byte("servicenow:\n  url: "+srv.URL+"/api\n  username: u\n  password: p\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	resetFlags(rootCmd)

	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	rootCmd.SetArgs(append([]string{"--config", cfg, "--nocolor"}, args...))
	err := rootCmd.Execute()
	w.Close()
	os.Stdout = stdout
	out, _ := io.ReadAll(r)
	return string(out), err
}

// resetFlags restores every flag to its default so state doesn't leak
// between tests that reuse the global command tree.
func resetFlags(c *cobra.Command) {
	reset := func(f *pflag.Flag) {
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			sv.Replace(nil)
		} else {
			f.Value.Set(f.DefValue)
		}
		f.Changed = false
	}
	c.Flags().VisitAll(reset)
	c.PersistentFlags().VisitAll(reset)
	for _, sub := range c.Commands() {
		resetFlags(sub)
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "change.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func decodeBody(t *testing.T, body string) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("request body is not valid JSON: %v\n%s", err, body)
	}
	return m
}

func TestGetChange(t *testing.T) {
	out, err := run(t, &fakeServiceNow{t: t}, "get", "change", "CHG0000001")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Patch servers") {
		t.Errorf("report output missing description:\n%s", out)
	}
}

func TestGetChangeRaw(t *testing.T) {
	out, err := run(t, &fakeServiceNow{t: t}, "get", "change", "CHG0000001", "-o", "raw")
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid([]byte(out)) {
		t.Errorf("raw output is not JSON: %q", out)
	}
}

func TestChangeNotFound(t *testing.T) {
	for _, number := range []string{"CHG0000404", "CHGMISMATCH"} {
		for _, args := range [][]string{
			{"get", "change", number},
			{"approve", "change", number},
			{"cancel", "change", number, "--reason", "x"},
			{"close", "change", number, "--note", "x", "--code", "y"},
			{"reject", "change", number, "--comment", "x"},
			{"schedule", "change", number, "--end", "tomorrow"},
			{"implement", "change", number, "--end", "tomorrow"},
		} {
			fake := &fakeServiceNow{t: t}
			_, err := run(t, fake, args...)
			if err == nil || !strings.Contains(err.Error(), "not found") {
				t.Errorf("%v: err = %v, want not found", args, err)
			}
			if w := fake.writes(); len(w) != 0 {
				t.Errorf("%v: modified a record that wasn't requested: %+v", args, w)
			}
		}
	}
}

func TestApproveChange(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	out, err := run(t, fake, "approve", "change", "CHG0000001", "-o", "raw")
	if err != nil {
		t.Fatal(err)
	}
	w := fake.writes()
	if len(w) != 1 || w[0].Method != "PATCH" || w[0].Path != "/api/sn_chg_rest/v1/change/abc123/approvals" {
		t.Fatalf("unexpected writes: %+v", w)
	}
	if decodeBody(t, w[0].Body)["state"] != "approved" {
		t.Errorf("body = %s", w[0].Body)
	}
	// raw output must be the result of the update, not the initial lookup
	if !strings.Contains(out, `"state":"updated"`) {
		t.Errorf("raw output is not the update response: %s", out)
	}
}

func TestFreeTextIsEscaped(t *testing.T) {
	text := "said \"no\"\nbecause \\ reasons"
	tests := []struct {
		args  []string
		path  string
		field string
	}{
		{[]string{"cancel", "change", "CHG0000001", "--reason", text}, "/api/sn_chg_rest/v1/change/normal/abc123", "work_notes"},
		{[]string{"close", "change", "CHG0000001", "--code", "successful", "--note", text}, "/api/sn_chg_rest/v1/change/normal/abc123", "close_notes"},
		{[]string{"reject", "change", "CHG0000001", "--comment", text}, "/api/sn_chg_rest/v1/change/abc123/approvals", "comments"},
	}
	for _, tt := range tests {
		fake := &fakeServiceNow{t: t}
		if _, err := run(t, fake, tt.args...); err != nil {
			t.Fatalf("%v: %v", tt.args, err)
		}
		w := fake.writes()
		if len(w) != 1 || w[0].Path != tt.path {
			t.Fatalf("%v: unexpected writes %+v", tt.args, w)
		}
		if got := decodeBody(t, w[0].Body)[tt.field]; got != text {
			t.Errorf("%v: %s = %q, want %q", tt.args, tt.field, got, text)
		}
	}
}

func TestRejectChangePrintsResult(t *testing.T) {
	out, err := run(t, &fakeServiceNow{t: t}, "reject", "change", "CHG0000001", "--comment", "no")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "updated") {
		t.Errorf("reject printed no result: %q", out)
	}
}

func TestCloseChangeRequiresArgument(t *testing.T) {
	if _, err := run(t, &fakeServiceNow{t: t}, "close", "change", "--note", "x", "--code", "y"); err == nil {
		t.Fatal("expected an error without a change number")
	}
}

func TestInvalidOutputRejectedBeforeAnyRequest(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	_, err := run(t, fake, "approve", "change", "CHG0000001", "-o", "jsno")
	if err == nil || !strings.Contains(err.Error(), "invalid output type") {
		t.Fatalf("err = %v", err)
	}
	if len(fake.requests) != 0 {
		t.Errorf("requests made despite invalid output: %+v", fake.requests)
	}
}

func TestScheduleChange(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	_, err := run(t, fake, "schedule", "change", "CHG0000001", "--start", "2030-01-02 10:00:00", "--end", "2030-01-02 12:30:00")
	if err != nil {
		t.Fatal(err)
	}
	w := fake.writes()
	if len(w) != 1 || w[0].Path != "/api/sn_chg_rest/v1/change/normal/abc123" {
		t.Fatalf("unexpected writes %+v", w)
	}
	body := decodeBody(t, w[0].Body)
	if body["start_date"] != "2030-01-02 10:00:00" || body["end_date"] != "2030-01-02 12:30:00" {
		t.Errorf("body = %v", body)
	}
}

func TestTimeWindowValidation(t *testing.T) {
	for _, tt := range []struct {
		args []string
		want string
	}{
		{[]string{"schedule", "change", "CHG0000001", "--start", "2030-01-02 10:00:00", "--end", "2030-01-01 10:00:00"}, "is before --start"},
		{[]string{"implement", "change", "CHG0000001", "--start", "", "--end", "tomorrow"}, "--start must not be empty"},
		{[]string{"schedule", "change", "CHG0000001"}, `required flag(s) "end" not set`},
	} {
		fake := &fakeServiceNow{t: t}
		_, err := run(t, fake, tt.args...)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%v: err = %v, want %q", tt.args, err, tt.want)
		}
		if w := fake.writes(); len(w) != 0 {
			t.Errorf("%v: unexpected writes %+v", tt.args, w)
		}
	}
}

func TestImplementChange(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	if _, err := run(t, fake, "implement", "change", "CHG0000001", "--start", "2030-01-02 10:00:00", "--end", "2030-01-02 11:00:00"); err != nil {
		t.Fatal(err)
	}
	w := fake.writes()
	if len(w) != 1 || w[0].Path != "/api/now/v1/table/change_request/abc123" {
		t.Fatalf("unexpected writes %+v", w)
	}
	if body := decodeBody(t, w[0].Body); body["work_start"] != "2030-01-02 10:00:00" || body["work_end"] != "2030-01-02 11:00:00" {
		t.Errorf("body = %v", body)
	}
}

func TestAddChangeResolvesAssignmentGroup(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	file := writeFile(t, "shortDescription: Patch servers\nassignmentGroup: Ops\n")
	if _, err := run(t, fake, "add", "change", "-f", file, "-t", "standard"); err != nil {
		t.Fatal(err)
	}
	w := fake.writes()
	if len(w) != 1 || w[0].Method != "POST" || w[0].Path != "/api/sn_chg_rest/v1/change/standard" {
		t.Fatalf("unexpected writes %+v", w)
	}
	body := decodeBody(t, w[0].Body)
	if body["assignment_group"] != "grp1" || body["short_description"] != "Patch servers" {
		t.Errorf("body = %v", body)
	}
}

func TestAddChangeErrors(t *testing.T) {
	for _, tt := range []struct {
		name string
		file string
		want string
	}{
		{"numeric group", "assignmentGroup: 42\n", "assignment group must be a group name"},
		{"unknown group", "assignmentGroup: Nope\n", `assignment group "Nope" not found`},
		{"missing file", "", "no such file"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "missing.yaml")
			if tt.file != "" {
				path = writeFile(t, tt.file)
			}
			fake := &fakeServiceNow{t: t}
			_, err := run(t, fake, "add", "change", "-f", path)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want %q", err, tt.want)
			}
			if w := fake.writes(); len(w) != 0 {
				t.Errorf("unexpected writes %+v", w)
			}
		})
	}
}

func TestEditChange(t *testing.T) {
	fake := &fakeServiceNow{t: t}
	file := writeFile(t, "shortDescription: Updated\n")
	if _, err := run(t, fake, "edit", "change", "CHG0000001", "-f", file); err != nil {
		t.Fatal(err)
	}
	w := fake.writes()
	if len(w) != 1 || w[0].Method != "PATCH" || w[0].Path != "/api/sn_chg_rest/v1/change/normal/abc123" {
		t.Fatalf("unexpected writes %+v", w)
	}
	if decodeBody(t, w[0].Body)["short_description"] != "Updated" {
		t.Errorf("body = %s", w[0].Body)
	}
}

func TestServiceNowErrorIsReported(t *testing.T) {
	fake := &fakeServiceNow{t: t, writeStatus: http.StatusForbidden}
	_, err := run(t, fake, "approve", "change", "CHG0000001")
	want := "ServiceNow returned 403 Forbidden: Operation failed: State transition not allowed"
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

func TestGetGroup(t *testing.T) {
	out, err := run(t, &fakeServiceNow{t: t}, "get", "group", "Ops", "-o", "prettyjson")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "grp1") {
		t.Errorf("output = %s", out)
	}
	if _, err := run(t, &fakeServiceNow{t: t}, "get", "group", "Nope"); err == nil {
		t.Error("expected error for unknown group")
	}
}
