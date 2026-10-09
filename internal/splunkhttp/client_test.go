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
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientTLSVerification(t *testing.T) {
	testCases := []struct {
		name                       string
		endpoint                   string
		insecureSkipVerify         bool
		serverName                 string
		expectedInsecureSkipVerify bool
		expectedServerName         string
	}{
		{name: "IPv4 loopback no server name", endpoint: "https://127.0.0.1:8089", expectedInsecureSkipVerify: true},
		{name: "IPv4 loopback range no server name", endpoint: "https://127.12.34.56:8089", expectedInsecureSkipVerify: true},
		{name: "IPv6 loopback no server name", endpoint: "https://[::1]:8089", expectedInsecureSkipVerify: true},
		{name: "IPv4 loopback with server name", endpoint: "https://127.0.0.1:8089", serverName: "splunkd.example.com", expectedServerName: "splunkd.example.com"},
		{name: "IPv6 loopback with server name", endpoint: "https://[::1]:8089", serverName: "splunkd.example.com", expectedServerName: "splunkd.example.com"},
		{name: "loopback over HTTP", endpoint: "http://127.0.0.1:8089", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "localhost hostname", endpoint: "https://localhost:8089", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "loopback-looking hostname", endpoint: "https://127.0.0.1.example.com:8089", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "loopback user info", endpoint: "https://127.0.0.1@evil.example:8089", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "remote IP", endpoint: "https://192.0.2.1:8089", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "remote hostname", endpoint: "https://splunkd.example.com:8089", expectedInsecureSkipVerify: false},
		{name: "malformed URL", endpoint: "https://%gh", serverName: "splunkd.example.com", expectedInsecureSkipVerify: false},
		{name: "explicit override wins over server name", endpoint: "https://127.0.0.1:8089", insecureSkipVerify: true, serverName: "splunkd.example.com", expectedInsecureSkipVerify: true},
		{name: "explicit override on remote host", endpoint: "https://splunkd.example.com:8089", insecureSkipVerify: true, expectedInsecureSkipVerify: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(tc.endpoint, time.Second, tc.insecureSkipVerify, tc.serverName)
			assert.Equal(t, time.Second, client.Timeout)
			assert.Equal(t, tc.expectedInsecureSkipVerify, skipsTLSVerification(client))
			assert.Equal(t, tc.expectedServerName, tlsServerName(client))
		})
	}
}

func TestNewClientClonesDefaultTransport(t *testing.T) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	require.True(t, ok)

	client := NewClient("https://127.0.0.1:8089", 0, false, "")
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotSame(t, defaultTransport, transport)
	assert.False(t, skipsTLSVerification(&http.Client{Transport: defaultTransport}))
}

func skipsTLSVerification(client *http.Client) bool {
	transport, ok := client.Transport.(*http.Transport)
	return ok && transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify
}

func tlsServerName(client *http.Client) string {
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil {
		return ""
	}
	return transport.TLSClientConfig.ServerName
}
