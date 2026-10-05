// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunktcp is a UF-native TCP receiver wrapper. Its confmap config
// carries the fields of a resolved [tcp://...] stanza and rebuilds the
// underlying tcpreceiver by delegating to tabuilder.
//
// The type is registered so a .conf config source can emit it. It is
// undocumented and not intended to be written by hand in YAML.
package splunktcp

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// TypeStr is the UF-native TCP receiver type.
const TypeStr = "splunk_tcp"

// Config is the UF-native, confmap-serializable TCP receiver config. Stanza
// params with no modeled field land in Extra and are passed through to
// tabuilder. Props and Transforms carry the shared props/transforms set,
// matching splunk_inputs.
type Config struct {
	Extra         map[string]string `mapstructure:",remain"`
	ListenAddress string            `mapstructure:"listen_address"`
	Index         string            `mapstructure:"index"`
	Source        string            `mapstructure:"source"`
	Sourcetype    string            `mapstructure:"sourcetype"`
	Host          string            `mapstructure:"host"`
	Props         []conf.Prop       `mapstructure:"props"`
	Transforms    []conf.Transform  `mapstructure:"transforms"`
	Port          int               `mapstructure:"port"`
}

// Validate requires a usable port. UF's port-only stanza form ([tcp://5514])
// leaves the address empty, which means every interface, so an empty
// ListenAddress is valid; a zero port is not, since it would bind an arbitrary
// one.
func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("splunk_tcp: port must be between 1 and 65535, got %d", c.Port)
	}
	return nil
}

type splunkTCP struct{}

func (splunkTCP) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, c.input(), c.Transforms, c.Props, set.TelemetrySettings)
}

// input rebuilds the resolved stanza this config was emitted from, which is what
// tabuilder dispatches on. An empty ListenAddress becomes 0.0.0.0, matching UF's
// port-only stanza form.
func (c *Config) input() conf.Input {
	addr := c.ListenAddress
	if addr == "" {
		addr = "0.0.0.0"
	}
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name: "tcp://" + net.JoinHostPort(addr, strconv.Itoa(c.Port)),
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

// NewFactory returns the UF-native TCP receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkTCP{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
