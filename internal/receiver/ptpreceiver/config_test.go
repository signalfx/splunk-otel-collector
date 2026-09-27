// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	config := createDefaultConfig().(*Config)
	require.NoError(t, config.Validate())
	for _, test := range []struct {
		name   string
		change func(*Config)
	}{
		{"relative socket", func(c *Config) { c.SocketPath = "ptp4l" }},
		{"empty pmc", func(c *Config) { c.PMCPath = "" }},
		{"invalid domain", func(c *Config) { c.DomainNumber = 256 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := *config
			test.change(&invalid)
			require.Error(t, invalid.Validate())
		})
	}
}
