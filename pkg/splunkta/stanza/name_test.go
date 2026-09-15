// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package stanza

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		name   string
		parsed Name
	}{
		{name: "WinEventLog://DFS some name", parsed: Name{Kind: "WinEventLog", Target: "DFS some name"}},
		{name: "monitor:///var/log/splunk/*.log", parsed: Name{Kind: "monitor", Target: "/var/log/splunk/*.log"}},
		{name: `monitor://C:\Program Files\Splunk\splunkd.log`, parsed: Name{Kind: "monitor", Target: `C:\Program Files\Splunk\splunkd.log`}},
		{name: "tcp://:9997", parsed: Name{Kind: "tcp", Target: ":9997"}},
		{name: "UDP:514", parsed: Name{Kind: "UDP", Target: "514"}},
		{name: "my_script", parsed: Name{Target: "my_script"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseName(test.name)
			require.NoError(t, err)
			require.Equal(t, test.parsed, parsed)
		})
	}
}

func TestParseNameRejectsEmptyKind(t *testing.T) {
	_, err := ParseName("://target")
	require.ErrorIs(t, err, errEmptyKind)
}

func TestParseOutputName(t *testing.T) {
	tests := []struct {
		name   string
		parsed Name
	}{
		{name: "httpout", parsed: Name{Kind: "httpout"}},
		{name: "tcpout:primary", parsed: Name{Kind: "tcpout", Target: "primary"}},
		{name: "s2s://primary", parsed: Name{Kind: "s2s", Target: "primary"}},
		{name: "tcpout-server://splunk:9997", parsed: Name{Kind: "tcpout-server", Target: "splunk:9997"}},
		{name: "S2S:Primary", parsed: Name{Kind: "S2S", Target: "Primary"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseOutputName(test.name)
			require.NoError(t, err)
			require.Equal(t, test.parsed, parsed)
		})
	}
}

func TestParseOutputNameRejectsEmptyKind(t *testing.T) {
	for _, name := range []string{"", "://target", ":target"} {
		_, err := ParseOutputName(name)
		require.ErrorIs(t, err, errEmptyKind)
	}
}

func TestListenAddress(t *testing.T) {
	require.Equal(t, ":9997", ListenAddress("9997"))
	require.Equal(t, ":9997", ListenAddress(":9997"))
	require.Equal(t, "127.0.0.1:9997", ListenAddress("127.0.0.1:9997"))
	require.Equal(t, "[::1]:9997", ListenAddress("[::1]:9997"))
}
