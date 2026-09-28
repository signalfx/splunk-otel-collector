// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package ptpreceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	config := createDefaultConfig().(*Config)
	require.NoError(t, config.Validate())
	for _, test := range []struct {
		change func(*Config)
		name   string
	}{
		{name: "relative socket", change: func(c *Config) { c.SocketPath = "ptp4l" }},
		{name: "empty socket", change: func(c *Config) { c.SocketPath = "" }},
		{name: "empty pmc", change: func(c *Config) { c.PMCPath = "" }},
		{name: "invalid domain", change: func(c *Config) { c.DomainNumber = 256 }},
		{name: "negative domain", change: func(c *Config) { c.DomainNumber = -1 }},
		{name: "invalid interval", change: func(c *Config) { c.CollectionInterval = -time.Second }},
	} {
		t.Run(test.name, func(t *testing.T) {
			invalid := *config
			test.change(&invalid)
			require.Error(t, invalid.Validate())
		})
	}
}
