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

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	persistentconn "github.com/DrakeW/splunk-persistentconn"
)

func TestProxyForwardsRequestAndReturnsUpstreamResponse(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotBody, gotHeader, gotConnection string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("format")
		gotHeader = r.Header.Get("X-Request-ID")
		gotConnection = r.Header.Get("Connection")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upstream request body: %v", err)
			return
		}
		gotBody = string(body)
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"accepted":true}`)
	}))
	defer upstream.Close()

	handler := newProxyHandler(upstream.URL, upstream.Client())
	response, err := handler(persistentconn.Request{
		Method:  http.MethodPost,
		Path:    "v1/traces",
		Query:   map[string]string{"format": "json output"},
		Headers: map[string]string{"X-Request-ID": "request-123", "Connection": "keep-alive"},
		Payload: `{"resourceSpans":[]}`,
	})
	if err != nil {
		t.Fatalf("proxy handler returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("upstream method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/v1/traces" {
		t.Errorf("upstream path = %q, want %q", gotPath, "/v1/traces")
	}
	if gotQuery != "json output" {
		t.Errorf("upstream query format = %q, want %q", gotQuery, "json output")
	}
	if gotHeader != "request-123" {
		t.Errorf("upstream X-Request-ID = %q, want %q", gotHeader, "request-123")
	}
	if gotConnection != "" {
		t.Errorf("hop-by-hop Connection header was forwarded: %q", gotConnection)
	}
	if gotBody != `{"resourceSpans":[]}` {
		t.Errorf("upstream body = %q, want %q", gotBody, `{"resourceSpans":[]}`)
	}
	if response.StatusCode != http.StatusAccepted {
		t.Errorf("response status = %d, want %d", response.StatusCode, http.StatusAccepted)
	}
	if response.Body != `{"accepted":true}` {
		t.Errorf("response body = %q, want %q", response.Body, `{"accepted":true}`)
	}
}

func TestProxyReturnsBadGatewayWhenUpstreamIsUnavailable(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := upstream.URL
	upstream.Close()

	handler := newProxyHandler(endpoint, &http.Client{})
	response, err := handler(persistentconn.Request{Method: http.MethodGet, Path: "/v1/metrics"})
	if err != nil {
		t.Fatalf("proxy handler returned error: %v", err)
	}
	if response.StatusCode != http.StatusBadGateway {
		t.Errorf("response status = %d, want %d", response.StatusCode, http.StatusBadGateway)
	}
	if !strings.Contains(response.Body, "upstream request failed") {
		t.Errorf("response body = %q, want upstream failure message", response.Body)
	}
}
