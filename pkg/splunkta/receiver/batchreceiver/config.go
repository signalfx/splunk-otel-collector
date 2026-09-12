// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package batchreceiver implements the batch receiver.
package batchreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

// Config holds the configuration for the batch receiver.
type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}
