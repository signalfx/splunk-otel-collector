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
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestConfigWithDefaults(t *testing.T) {
	c := Config{}.withDefaults()
	if c.Image != defaultImage || c.Password != defaultPassword || c.StartupTimeout != 5*time.Minute {
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

	// A malformed _time is ignored, leaving the zero value.
	if got := resultToRecord(map[string]any{"_time": "not-a-time"}); !got.Time.IsZero() {
		t.Errorf("bad _time should stay zero, got %v", got.Time)
	}
}

func TestHECAndMgmtURL(t *testing.T) {
	s := New(Config{Token: "11111111-1111-1111-1111-111111111111"})
	s.host = "example.com"
	s.hecPort = "8088"
	s.mgmtPort = "8089"

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
	ts := httptest.NewTLSServer(h)
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	s := New(Config{Password: "pw", Token: "tok"})
	s.host = u.Hostname()
	s.mgmtPort = u.Port()
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
