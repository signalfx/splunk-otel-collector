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

package splunkhttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewClient returns an HTTP client for a Splunk management endpoint. Splunk
// modular inputs commonly receive a loopback IP in server_uri even when the
// splunkd certificate only identifies the host by DNS name, so certificate
// verification is automatically skipped for literal loopback HTTPS endpoints.
// Verification remains enabled for hostnames and non-loopback addresses unless
// insecureSkipVerify is explicitly set.
func NewClient(endpoint string, timeout time.Duration, insecureSkipVerify bool) *http.Client {
	client := &http.Client{Timeout: timeout}
	if insecureSkipVerify || isLoopbackHTTPS(endpoint) {
		client.Transport = insecureTransport()
	}
	return client
}

func isLoopbackHTTPS(endpoint string) bool {
	parsed, err := url.Parse(endpoint)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}
	ip := net.ParseIP(parsed.Hostname())
	return ip != nil && ip.IsLoopback()
}

func insecureTransport() *http.Transport {
	transport, ok := http.DefaultTransport.(*http.Transport)
	// The cast can fail if an application or test replaces the default transport
	// with another http.RoundTripper implementation.
	if ok {
		transport = transport.Clone()
	} else {
		transport = &http.Transport{}
	}

	tlsConfig := transport.TLSClientConfig
	// The default transport leaves TLSClientConfig nil to use Go's TLS defaults.
	if tlsConfig == nil {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	} else {
		tlsConfig = tlsConfig.Clone()
	}
	tlsConfig.InsecureSkipVerify = true
	transport.TLSClientConfig = tlsConfig
	return transport
}
