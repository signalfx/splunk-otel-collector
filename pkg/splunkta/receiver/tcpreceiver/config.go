// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package tcpreceiver implements the TCP receiver.
package tcpreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

// Config holds the configuration for the TCP receiver.
type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}

// Validate validates the Config.
func (cfg *Config) Validate() error {
	_, err := stanza.ParseName(cfg.Input.Configuration.Stanza.Name)
	return err
}
