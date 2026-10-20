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
// splunkd certificate only identifies the host by DNS name, so plain
// certificate verification fails for literal loopback HTTPS endpoints. There
// is no reliable hostname to verify against in that case: server_host from
// the modular input protocol is not guaranteed to match a DNS SAN on the
// certificate either. So when insecureSkipVerify is nil (not explicitly
// configured by the caller), verification is skipped for literal loopback
// HTTPS endpoints the same way it would be for an explicit true, and remains
// enabled for every other endpoint. An explicit insecureSkipVerify value is
// always honored as-is, including on loopback endpoints, so a caller that
// provisions a certificate with loopback SANs and explicitly sets
// insecureSkipVerify to false gets real verification.
func NewClient(endpoint string, timeout time.Duration, insecureSkipVerify *bool) *http.Client {
	client := &http.Client{Timeout: timeout}
	skipVerify := insecureSkipVerify != nil && *insecureSkipVerify
	fallbackToInsecure := insecureSkipVerify == nil && isLoopbackHTTPS(endpoint)
	if skipVerify || fallbackToInsecure {
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
	transport := cloneDefaultTransport()
	tlsConfig := cloneTLSConfig(transport)
	tlsConfig.InsecureSkipVerify = true
	transport.TLSClientConfig = tlsConfig
	return transport
}

func cloneDefaultTransport() *http.Transport {
	transport, ok := http.DefaultTransport.(*http.Transport)
	// The cast can fail if an application or test replaces the default transport
	// with another http.RoundTripper implementation.
	if ok {
		return transport.Clone()
	}
	return &http.Transport{}
}

func cloneTLSConfig(transport *http.Transport) *tls.Config {
	// The default transport leaves TLSClientConfig nil to use Go's TLS defaults.
	if transport.TLSClientConfig == nil {
		return &tls.Config{MinVersion: tls.VersionTLS12}
	}
	return transport.TLSClientConfig.Clone()
}
