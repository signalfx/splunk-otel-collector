// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package scriptreceiver implements the script receiver.
package scriptreceiver

import "github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"

// Config holds the configuration for the script receiver.
type Config struct {
	conf.Input `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
}
