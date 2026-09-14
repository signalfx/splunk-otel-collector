// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package wineventlogreceiver implements the Windows Event Log receiver.
package wineventlogreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// Config holds the configuration for the Windows Event Log receiver.
type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}
