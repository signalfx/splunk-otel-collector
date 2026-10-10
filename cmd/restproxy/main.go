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
	"os"
	"strings"
	"time"

	splunkproxy "github.com/signalfx/splunk-otel-collector/internal/splunkproxy"
)

const collectorEndpoint = "http://127.0.0.1:4318"

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	server := splunkproxy.NewServer()
	proxy := newProxyHandler(collectorEndpoint, &http.Client{Timeout: 30 * time.Second})
	server.Handle(".*", proxy, "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD")
	return server.Run()
}

func newProxyHandler(endpoint string, client *http.Client) splunkproxy.Handler {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		panic(fmt.Errorf("parse collector endpoint: %w", err))
	}

	return func(in splunkproxy.Request) (splunkproxy.Response, error) {
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

		request, err := http.NewRequest(in.Method, target.String(), bytes.NewReader(in.Body))
		if err != nil {
			return splunkproxy.Response{}, fmt.Errorf("create upstream request: %w", err)
		}
		connectionTokens := connectionHeaderTokens(in.Headers)
		for name, values := range in.Headers {
			if !isHopByHopHeader(name) {
				if _, namedByConnection := connectionTokens[http.CanonicalHeaderKey(name)]; !namedByConnection {
					for _, value := range values {
						request.Header.Add(name, value)
					}
				}
			}
		}
		request.Host = target.Host

		response, err := client.Do(request)
		if err != nil {
			return splunkproxy.Response{
				StatusCode: http.StatusBadGateway,
				Body:       []byte(fmt.Sprintf("upstream request failed: %v", err)),
			}, nil
		}
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return splunkproxy.Response{}, fmt.Errorf("read upstream response: %w", err)
		}
		responseHeaders := make(http.Header)
		responseConnectionTokens := connectionHeaderTokens(response.Header)
		for name, values := range response.Header {
			if !isHopByHopHeader(name) {
				if _, namedByConnection := responseConnectionTokens[http.CanonicalHeaderKey(name)]; namedByConnection {
					continue
				}
				responseHeaders[name] = append([]string(nil), values...)
			}
		}
		return splunkproxy.Response{StatusCode: response.StatusCode, Body: responseBody, Headers: responseHeaders}, nil
	}
}

func connectionHeaderTokens(headers http.Header) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, value := range headers.Values("Connection") {
		for token := range strings.SplitSeq(value, ",") {
			tokens[http.CanonicalHeaderKey(strings.TrimSpace(token))] = struct{}{}
		}
	}
	return tokens
}

func isHopByHopHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}
