// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package monitorreceiver implements the monitor receiver.
package monitorreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// Config holds the configuration for the monitor receiver.
type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}
