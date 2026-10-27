// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package splunk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConfigWithDefaults(t *testing.T) {
	clearParityEnv(t)
	c := Config{}.withDefaults()
	if c.Image != defaultImage || c.User != defaultUser || c.Password != defaultPassword ||
		c.StartupTimeout != 5*time.Minute {
		t.Errorf("defaults = %+v", c)
	}
	if len(c.Token) < 36 {
		t.Errorf("default token must be >= 36 chars (UF httpout requirement), got %q", c.Token)
	}
	custom := Config{Image: "img", Password: "pw", Token: "tok", StartupTimeout: time.Second}.withDefaults()
	if custom.Image != "img" || custom.Password != "pw" || custom.Token != "tok" || custom.StartupTimeout != time.Second {
		t.Errorf("custom values overridden: %+v", custom)
	}
}

func TestStr(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{"hello", "hello"},
		{[]any{"a", "b"}, "a"}, // multivalued: first value
		{[]any{}, "[]"},        // empty multivalue falls through to fmt.Sprint
		{nil, ""},
		{42, "42"},
	}
	for _, c := range cases {
		if got := str(c.in); got != c.want {
			t.Errorf("str(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResultToRecord(t *testing.T) {
	r := resultToRecord(map[string]any{
		"_raw":       "line",
		"host":       "h",
		"source":     "s",
		"sourcetype": "st",
		"index":      "i",
		"_time":      "2026-09-23T10:00:00Z",
	})
	if r.Raw != "line" || r.Host != "h" || r.Source != "s" || r.Sourcetype != "st" || r.Index != "i" {
		t.Errorf("record = %+v", r)
	}
	if r.Time.IsZero() {
		t.Error("_time not parsed")
	}

	if len(r.Fields) != 0 {
		t.Errorf("typed fields should not be repeated in Fields: %+v", r.Fields)
	}

	// A malformed _time is ignored, leaving the zero value.
	if got := resultToRecord(map[string]any{"_time": "not-a-time"}); !got.Time.IsZero() {
		t.Errorf("bad _time should stay zero, got %v", got.Time)
	}
}

// TestResultToRecordFields checks the dynamic search fields a case can select
// land in Record.Fields, and that Splunk's volatile internals do not.
func TestResultToRecordFields(t *testing.T) {
	r := resultToRecord(map[string]any{
		"_raw":        "line",
		"host":        "h",
		"punct":       "--_",
		"linecount":   "1",
		"date_wday":   "wednesday",
		"multivalued": []any{"first", "second"},
		"_bkt":        "main~0~ABC",
		"_cd":         "0:1",
		"_indextime":  "1790000000",
	})
	want := map[string]string{
		"punct":       "--_",
		"linecount":   "1",
		"date_wday":   "wednesday",
		"multivalued": "first",
	}
	if !reflect.DeepEqual(r.Fields, want) {
		t.Errorf("Fields = %+v, want %+v", r.Fields, want)
	}
}

func TestHECAndMgmtURL(t *testing.T) {
	s := New(Config{Token: "11111111-1111-1111-1111-111111111111"})
	s.hecBase = "https://example.com:8088"
	s.mgmtBase = "https://example.com:8089"

	hec := s.HEC()
	if hec.Endpoint != "https://example.com:8088" || hec.Token != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("HEC = %+v", hec)
	}
	if s.mgmtURL() != "https://example.com:8089" {
		t.Errorf("mgmtURL = %q", s.mgmtURL())
	}
}

// newTestSplunk returns a Splunk backend whose management API is served by h
// over TLS, so the REST client paths can run without Docker.
func newTestSplunk(t *testing.T, h http.Handler) *Splunk {
	t.Helper()
	clearParityEnv(t)
	ts := httptest.NewTLSServer(h)
	t.Cleanup(ts.Close)
	s := New(Config{Password: "pw", Token: "tok"})
	s.mgmtBase = ts.URL
	return s
}

// searchMux serves the three-step search job flow (create, poll, read events)
// and records the SPL it was asked to run.
func searchMux(capturedSPL *string, eventsBody string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/services/search/jobs", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		*capturedSPL = r.FormValue("search")
		_, _ = w.Write([]byte(`{"sid":"S1"}`))
	})
	mux.HandleFunc("/services/search/jobs/S1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"entry":[{"content":{"isDone":true}}]}`))
	})
	mux.HandleFunc("/services/search/jobs/S1/events", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(eventsBody))
	})
	return mux
}

func TestSearch(t *testing.T) {
	var spl string
	s := newTestSplunk(t, searchMux(&spl, `{"results":[{"_raw":"hi","host":"h","index":"i"}]}`))

	recs, err := s.Search(context.Background(), "search index=i")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(recs) != 1 || recs[0].Raw != "hi" || recs[0].Host != "h" || recs[0].Index != "i" {
		t.Errorf("records = %+v", recs)
	}
	if spl != "search index=i" {
		t.Errorf("posted SPL = %q", spl)
	}
}

func TestClean(t *testing.T) {
	var spl string
	s := newTestSplunk(t, searchMux(&spl, `{"results":[]}`))

	if err := s.Clean(context.Background(), "index=i"); err != nil {
		t.Fatalf("Clean: %v", err)
	}
	if !strings.Contains(spl, "| delete") {
		t.Errorf("Clean SPL missing delete: %q", spl)
	}
}

func TestSearchNoSid(t *testing.T) {
	s := newTestSplunk(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`)) // no sid
	}))
	if _, err := s.Search(context.Background(), "x"); err == nil {
		t.Fatal("expected error when no sid is returned")
	}
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"created", http.StatusCreated, false},
		{"already exists", http.StatusConflict, false},
		{"server error", http.StatusInternalServerError, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotName string
			mux := http.NewServeMux()
			mux.HandleFunc("/services/data/indexes", func(w http.ResponseWriter, r *http.Request) {
				_ = r.ParseForm()
				gotName = r.FormValue("name")
				w.WriteHeader(tt.status)
			})
			s := newTestSplunk(t, mux)
			err := s.createIndex(context.Background(), "myidx")
			if (err != nil) != tt.wantErr {
				t.Fatalf("createIndex err = %v, wantErr %v", err, tt.wantErr)
			}
			if gotName != "myidx" {
				t.Errorf("index name = %q", gotName)
			}
		})
	}
}

func TestStopNilContainer(t *testing.T) {
	if err := New(Config{}).Stop(context.Background()); err != nil {
		t.Errorf("Stop with no container = %v, want nil", err)
	}
}

func TestConfigEnv(t *testing.T) {
	clearParityEnv(t)
	t.Setenv(EnvHECURL, "http://splunk:8088")
	t.Setenv(EnvMgmtURL, "https://splunk:8089")
	t.Setenv(EnvToken, "env-token")
	t.Setenv(EnvPassword, "env-pw")

	c := Config{}.withDefaults()
	if c.HECURL != "http://splunk:8088" || c.MgmtURL != "https://splunk:8089" ||
		c.Token != "env-token" || c.Password != "env-pw" {
		t.Errorf("env not resolved: %+v", c)
	}
}

func TestConfigExternalSkipsContainerDefaults(t *testing.T) {
	clearParityEnv(t)
	c := Config{HECURL: "http://splunk:8088", MgmtURL: "https://splunk:8089"}.withDefaults()
	if !c.external() {
		t.Fatal("external() = false")
	}
	if c.Password != "" || c.Token != "" || c.Image != "" || c.StartupTimeout != 0 {
		t.Errorf("container defaults leaked onto the external path: %+v", c)
	}
	if c.User != defaultUser {
		t.Errorf("User = %q, want %q", c.User, defaultUser)
	}
}

func TestConfigValidate(t *testing.T) {
	clearParityEnv(t)
	const hec, mgmt = "http://splunk:8088", "https://splunk:8089"

	tests := []struct {
		name    string
		wantErr string
		cfg     Config
	}{
		{"container defaults", "", Config{}},
		{"external complete", "", Config{HECURL: hec, MgmtURL: mgmt, Password: "pw", Token: defaultToken}},
		{"hec without mgmt", EnvMgmtURL, Config{HECURL: hec}},
		{"mgmt without hec", EnvHECURL, Config{MgmtURL: mgmt}},
		{"no password", EnvPassword, Config{HECURL: hec, MgmtURL: mgmt, Token: defaultToken}},
		{"no token", EnvToken, Config{HECURL: hec, MgmtURL: mgmt, Password: "pw"}},
		{"short token", ">= 36", Config{HECURL: hec, MgmtURL: mgmt, Password: "pw", Token: "short"}},
		{"zero guid", "all-zeros", Config{HECURL: hec, MgmtURL: mgmt, Password: "pw", Token: zeroGUID}},
		{"schemeless", "scheme", Config{HECURL: "splunk:8088", MgmtURL: mgmt, Password: "pw", Token: defaultToken}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.withDefaults().validate()
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.wantErr != "" && err == nil:
				t.Fatalf("want error mentioning %q, got nil", tt.wantErr)
			case tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr):
				t.Errorf("error %q does not mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestStartExternalSkipsContainer(t *testing.T) {
	clearParityEnv(t)
	var created []string
	mux := http.NewServeMux()
	mux.HandleFunc("/services/data/indexes", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if user, pass, _ := r.BasicAuth(); user != "svc" || pass != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		created = append(created, r.FormValue("name"))
		w.WriteHeader(http.StatusCreated)
	})
	ts := httptest.NewTLSServer(mux)
	t.Cleanup(ts.Close)

	s := New(Config{
		HECURL:   "http://hec.example.com:8088",
		MgmtURL:  ts.URL,
		User:     "svc",
		Password: "pw",
		Token:    defaultToken,
		Indexes:  []string{"parity_a", "parity_b"},
	})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if s.container != nil {
		t.Error("external Start created a container")
	}
	if hec := s.HEC(); hec.Endpoint != "http://hec.example.com:8088" || hec.Token != defaultToken {
		t.Errorf("HEC = %+v", hec)
	}
	if !reflect.DeepEqual(created, []string{"parity_a", "parity_b"}) {
		t.Errorf("created indexes = %v", created)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Errorf("Stop on an external backend: %v", err)
	}
}

func TestStartExternalTrimsTrailingSlash(t *testing.T) {
	clearParityEnv(t)
	s := New(Config{
		HECURL:   "https://hec.example.com:8088/",
		MgmtURL:  "https://mgmt.example.com:8089/",
		Password: "pw",
		Token:    defaultToken,
	})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if got := s.HEC().Endpoint; got != "https://hec.example.com:8088" {
		t.Errorf("HEC endpoint = %q", got)
	}
	if got := s.mgmtURL(); got != "https://mgmt.example.com:8089" {
		t.Errorf("mgmtURL = %q", got)
	}
}

func TestSearchUsesConfiguredUser(t *testing.T) {
	clearParityEnv(t)
	var gotUser, spl string
	inner := searchMux(&spl, `{"results":[]}`)
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, _, _ = r.BasicAuth()
		inner.ServeHTTP(w, r)
	}))
	t.Cleanup(ts.Close)

	s := New(Config{
		HECURL: "http://hec.example.com:8088", MgmtURL: ts.URL,
		User: "svc", Password: "pw", Token: defaultToken,
	})
	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if _, err := s.Search(context.Background(), "search index=i"); err != nil {
		t.Fatalf("Search: %v", err)
	}
	if gotUser != "svc" {
		t.Errorf("basic-auth user = %q, want svc", gotUser)
	}
}

func clearParityEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{EnvHECURL, EnvMgmtURL, EnvToken, EnvUser, EnvPassword, EnvImage} {
		t.Setenv(k, "")
	}
}
