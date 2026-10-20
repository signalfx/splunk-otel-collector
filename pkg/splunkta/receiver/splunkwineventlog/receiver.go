// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkwineventlog is a UF-native Windows Event Log receiver wrapper.
// Its confmap config carries the fields of a resolved [WinEventLog://...] stanza
// and rebuilds the underlying wineventlogreceiver by delegating to tabuilder.
//
// The type is registered so a .conf config source can emit it. It is
// undocumented and not intended to be written by hand in YAML.
package splunkwineventlog

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

// TypeStr is the UF-native Windows Event Log receiver type.
const TypeStr = "splunk_wineventlog"

// Config is the UF-native, confmap-serializable wineventlog receiver config.
// Stanza params with no modeled field land in Extra and are passed through to
// tabuilder. Props and Transforms carry the shared props/transforms set,
// matching splunk_inputs.
type Config struct {
	Extra        map[string]string `mapstructure:",remain"`
	EventLogName string            `mapstructure:"event_log_name"`
	Index        string            `mapstructure:"index,omitempty"`
	Source       string            `mapstructure:"source,omitempty"`
	Sourcetype   string            `mapstructure:"sourcetype,omitempty"`
	Host         string            `mapstructure:"host,omitempty"`
	Props        []conf.Prop       `mapstructure:"props,omitempty"`
	Transforms   []conf.Transform  `mapstructure:"transforms,omitempty"`
}

// Validate requires the channel name, which is the stanza target and has no
// meaningful default.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.EventLogName) == "" {
		return errors.New("splunk_wineventlog: event_log_name is required")
	}
	return nil
}

type splunkWineventlog struct{}

func (splunkWineventlog) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, c.input(), c.Transforms, c.Props, set.TelemetrySettings)
}

// input rebuilds the resolved stanza this config was emitted from, which is what
// tabuilder dispatches on.
func (c *Config) input() conf.Input {
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name: "WinEventLog://" + c.EventLogName,
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

// Factory is the UF-native wineventlog receiver factory. It maps its own
// [WinEventLog://...] stanzas, so the .conf provider models no input kind itself.
type Factory struct {
	receiver.Factory
}

// NewFactory returns the UF-native Windows Event Log receiver factory.
func NewFactory() Factory {
	return Factory{receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkWineventlog{}.CreateLogs, component.StabilityLevelAlpha),
	)}
}
