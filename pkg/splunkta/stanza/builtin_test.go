// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package stanza

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsBuiltinKind(t *testing.T) {
	tests := []struct {
		kind    string
		builtin bool
	}{
		{kind: "monitor", builtin: true},
		{kind: "MONITOR", builtin: true},
		{kind: "splunktcp-ssl", builtin: true},
		{kind: "WinEventLog"},
		{kind: ""},
	}

	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			require.Equal(t, test.builtin, IsBuiltinKind(test.kind))
		})
	}
}

func TestIsModularInput(t *testing.T) {
	tests := []struct {
		name     string
		modinput bool
	}{
		{name: "WinEventLog://Application", modinput: true},
		{name: "WinEventLog", modinput: true},
		{name: "monitor:///var/log/a"},
		{name: "monitor"},

		// The split is at the first colon, so a built-in stanza that separates
		// with a bare colon is not taken for a modular input.
		{name: "tcp:9997"},
		{name: "blacklist:/var/log/a"},

		// A non-built-in kind with a bare colon separator is not one either:
		// splunkd's check is for "://".
		{name: "WinEventLog:Application"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.modinput, IsModularInput(test.name))
		})
	}
}
