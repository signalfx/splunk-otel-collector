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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	persistentconn "github.com/DrakeW/splunk-persistentconn"
)

const collectorEndpoint = "http://127.0.0.1:4318"

func main() {
	server := persistentconn.NewServer()
	proxy := newProxyHandler(collectorEndpoint, &http.Client{Timeout: 30 * time.Second})
	server.Handle(".*", proxy, "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD")
	server.Run()
}

func newProxyHandler(endpoint string, client *http.Client) persistentconn.Handler {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		panic(fmt.Errorf("parse collector endpoint: %w", err))
	}

	return func(in persistentconn.Request) (persistentconn.Response, error) {
		path := in.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		target := *baseURL
		target.Path = strings.TrimRight(baseURL.Path, "/") + path
		query := target.Query()
		for key, value := range in.Query {
			query.Set(key, value)
		}
		target.RawQuery = query.Encode()

		request, err := http.NewRequest(in.Method, target.String(), bytes.NewReader([]byte(in.Payload)))
		if err != nil {
			return persistentconn.Response{}, fmt.Errorf("create upstream request: %w", err)
		}
		for name, value := range in.Headers {
			if !isHopByHopHeader(name) {
				request.Header.Set(name, value)
			}
		}
		request.Host = target.Host

		response, err := client.Do(request)
		if err != nil {
			return persistentconn.Response{
				StatusCode: http.StatusBadGateway,
				Body:       fmt.Sprintf("upstream request failed: %v", err),
			}, nil
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return persistentconn.Response{}, fmt.Errorf("read upstream response: %w", err)
		}
		return persistentconn.Response{StatusCode: response.StatusCode, Body: string(body)}, nil
	}
}

func isHopByHopHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}
