// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package batchreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}
