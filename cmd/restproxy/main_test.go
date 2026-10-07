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

	persistentconn "github.com/signalfx/splunk-otel-collector/internal/splunkproxy"
)

func TestProxyForwardsRequestAndReturnsUpstreamResponse(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotBody, gotHeader, gotConnection string
	var gotHeaderContentType string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.Query().Get("format")
		gotHeader = r.Header.Get("X-Request-ID")
		gotConnection = r.Header.Get("Connection")
		gotHeaderContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upstream request body: %v", err)
			return
		}
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"accepted":true}`)
	}))
	defer upstream.Close()

	handler := newProxyHandler(upstream.URL, upstream.Client())
	response, err := handler(persistentconn.Request{
		Method:  http.MethodPost,
		Path:    "v1/traces",
		Query:   map[string]string{"format": "json output"},
		Headers: http.Header{"X-Request-ID": {"request-123"}, "Connection": {"keep-alive, X-Hop-By-Hop"}, "X-Hop-By-Hop": {"remove-me"}, "Content-Type": {"application/x-protobuf"}},
		Body:    []byte(`{"resourceSpans":[]}`),
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
	if gotHeaderContentType != "application/x-protobuf" {
		t.Errorf("upstream Content-Type = %q, want %q", gotHeaderContentType, "application/x-protobuf")
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
	if string(response.Body) != `{"accepted":true}` {
		t.Errorf("response body = %q, want %q", response.Body, `{"accepted":true}`)
	}
	if response.Headers.Get("Content-Type") != "application/json" {
		t.Errorf("response Content-Type = %q, want application/json", response.Headers.Get("Content-Type"))
	}
}

func TestProxyUsesRestmapRouteAndForwardsBinaryPayload(t *testing.T) {
	wantBody := []byte{0, 1, 2, 255}
	var gotPath string
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/x-protobuf")
		_, _ = w.Write([]byte{0x0a, 0x03, 0x08, 0x96, 0x01})
	}))
	defer upstream.Close()

	handler := newProxyHandler(upstream.URL, upstream.Client())
	response, err := handler(persistentconn.Request{
		Method: http.MethodPost,
		Path:   "/v1/traces",
		Body:   wantBody,
	})
	if err != nil {
		t.Fatalf("proxy handler returned error: %v", err)
	}
	if gotPath != "/v1/traces" {
		t.Errorf("upstream path = %q, want /v1/traces", gotPath)
	}
	if string(gotBody) != string(wantBody) {
		t.Errorf("upstream body = %v, want %v", gotBody, wantBody)
	}
	wantResponse := []byte{0x0a, 0x03, 0x08, 0x96, 0x01}
	if string(response.Body) != string(wantResponse) {
		t.Errorf("response body = %v, want %v", response.Body, wantResponse)
	}
	if got := response.Headers.Get("Content-Type"); got != "application/x-protobuf" {
		t.Errorf("response Content-Type = %q, want application/x-protobuf", got)
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
	if !strings.Contains(string(response.Body), "upstream request failed") {
		t.Errorf("response body = %q, want upstream failure message", response.Body)
	}
}
