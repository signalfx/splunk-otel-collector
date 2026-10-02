// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkmonitor is a UF-native monitor receiver wrapper. Its confmap
// config carries the fields of a resolved [monitor://...] stanza
// (include/exclude/charset/truncate/event_breaker plus the resource attrs) and
// rebuilds the underlying monitorreceiver by delegating to tabuilder, so the
// data path is the one splunk_inputs already uses.
//
// The type is registered so a .conf config source can emit it. It is
// undocumented and not intended to be written by hand in YAML.
package splunkmonitor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// TypeStr is the UF-native monitor receiver type.
const TypeStr = "splunk_monitor"

// Config is the UF-native, confmap-serializable monitor config. Stanza params
// with no modeled field land in Extra and are passed through to tabuilder, so an
// unmodeled UF setting is neither dropped nor rejected by strict unmarshal.
// Props and Transforms carry the shared props/transforms set, matching
// splunk_inputs.
type Config struct {
	Extra        map[string]string `mapstructure:",remain"`
	Index        string            `mapstructure:"index"`
	Source       string            `mapstructure:"source"`
	Sourcetype   string            `mapstructure:"sourcetype"`
	Host         string            `mapstructure:"host"`
	Charset      string            `mapstructure:"charset"`
	EventBreaker string            `mapstructure:"event_breaker"`
	Include      []string          `mapstructure:"include"`
	Exclude      []string          `mapstructure:"exclude"`
	Props        []conf.Prop       `mapstructure:"props"`
	Transforms   []conf.Transform  `mapstructure:"transforms"`
	Truncate     int               `mapstructure:"truncate"`
}

// Validate rejects a config with no include path. The include path is the
// monitored target and there is no meaningful default, so without it the
// receiver would silently watch the wrong thing.
func (c *Config) Validate() error {
	if len(c.Include) == 0 {
		return errors.New("splunk_monitor: include is required")
	}
	for _, path := range c.Include {
		if strings.TrimSpace(path) == "" {
			return errors.New("splunk_monitor: include must not contain empty paths")
		}
	}
	if c.Truncate < 0 {
		return fmt.Errorf("splunk_monitor: truncate must not be negative, got %d", c.Truncate)
	}
	return nil
}

type splunkMonitor struct{}

func (splunkMonitor) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, c.input(), c.Transforms, c.Props, set.TelemetrySettings)
}

// input rebuilds the resolved stanza this config was emitted from, which is what
// tabuilder dispatches on.
func (c *Config) input() conf.Input {
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   "monitor://" + c.Include[0],
				Params: c.params(),
			},
		},
	}
}

func (c *Config) params() conf.Params {
	params := conf.InputParams(conf.ResourceAttrs{
		Index:      c.Index,
		Source:     c.Source,
		Sourcetype: c.Sourcetype,
		Host:       c.Host,
	}, c.Extra)

	if len(c.Exclude) > 0 {
		params = append(params, conf.Param{Name: "blacklist", Value: strings.Join(c.Exclude, "|")})
	}
	if c.Charset != "" {
		params = append(params, conf.Param{Name: "CHARSET", Value: c.Charset})
	}
	if c.EventBreaker != "" {
		params = append(params, conf.Param{Name: "EVENT_BREAKER", Value: c.EventBreaker})
	}
	if c.Truncate > 0 {
		params = append(params, conf.Param{Name: "TRUNCATE", Value: strconv.Itoa(c.Truncate)})
	}
	return params
}

// NewFactory returns the UF-native monitor receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkMonitor{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
