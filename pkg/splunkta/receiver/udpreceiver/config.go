// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package udpreceiver

import (
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

type Config struct {
	Input      conf.Input       `mapstructure:"-"`
	BaseDir    string           `mapstructure:"-"`
	Transforms []conf.Transform `mapstructure:"-"`
	Props      []conf.Prop      `mapstructure:"-"`
}

func (cfg *Config) Validate() error {
	_, err := stanza.ParseName(cfg.Input.Configuration.Stanza.Name)
	return err
}
