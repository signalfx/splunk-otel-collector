// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkbatch is a UF-native batch input receiver wrapper. Its confmap
// config carries the fields of a resolved [batch://...] stanza and rebuilds the
// underlying batchreceiver by delegating to tabuilder.
//
// The type is registered so a .conf config source can emit it. It is
// undocumented and not intended to be written by hand in YAML.
package splunkbatch

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// TypeStr is the UF-native batch input receiver type.
const TypeStr = "splunk_batch"

// Config is the UF-native, confmap-serializable batch receiver config. Stanza
// params with no modeled field land in Extra and are passed through to
// tabuilder. Props and Transforms carry the shared props/transforms set,
// matching splunk_inputs.
type Config struct {
	Extra      map[string]string `mapstructure:",remain"`
	FilePath   string            `mapstructure:"file_path"`
	Index      string            `mapstructure:"index"`
	Source     string            `mapstructure:"source"`
	Sourcetype string            `mapstructure:"sourcetype"`
	Host       string            `mapstructure:"host"`
	Props      []conf.Prop       `mapstructure:"props"`
	Transforms []conf.Transform  `mapstructure:"transforms"`
}

// Validate requires the file path, which is the stanza target and has no
// meaningful default.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.FilePath) == "" {
		return errors.New("splunk_batch: file_path is required")
	}
	return nil
}

type splunkBatch struct{}

func (splunkBatch) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, c.input(), c.Transforms, c.Props, set.TelemetrySettings)
}

// input rebuilds the resolved stanza this config was emitted from, which is what
// tabuilder dispatches on.
func (c *Config) input() conf.Input {
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name: "batch://" + c.FilePath,
				Params: conf.InputParams(conf.ResourceAttrs{
					Index:      c.Index,
					Source:     c.Source,
					Sourcetype: c.Sourcetype,
					Host:       c.Host,
				}, c.Extra),
			},
		},
	}
}

// NewFactory returns the UF-native batch input receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkBatch{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
