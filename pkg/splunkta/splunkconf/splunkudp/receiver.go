// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkudp

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const TypeStr = "splunk_udp"

type Config struct {
	ListenAddress string            `mapstructure:"listen_address"`
	Port          int               `mapstructure:"port"`
	Index         string            `mapstructure:"index"`
	Source        string            `mapstructure:"source"`
	Sourcetype    string            `mapstructure:"sourcetype"`
	Host          string            `mapstructure:"host"`
	Props         []conf.Prop       `mapstructure:"props"`
	Transforms    []conf.Transform  `mapstructure:"transforms"`
	Extra         map[string]string `mapstructure:",remain"`
}

func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535, got %d", c.Port)
	}
	return nil
}

type splunkUDP struct{}

func (splunkUDP) CreateDefaultConfig() component.Config { return &Config{} }

func (splunkUDP) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	target := c.ListenAddress
	if target == "" {
		target = "0.0.0.0"
	}
	if c.Port > 0 {
		target = target + ":" + strconv.Itoa(c.Port)
	}
	input := conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   "udp://" + target,
				Params: buildParams(c),
			},
		},
	}
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, input, c.Transforms, c.Props, set.TelemetrySettings)
}

func buildParams(c *Config) []conf.Param {
	params := []conf.Param{}
	for k, v := range map[string]string{
		"index":      c.Index,
		"source":     c.Source,
		"sourcetype": c.Sourcetype,
		"host":       c.Host,
	} {
		if v != "" {
			params = append(params, conf.Param{Name: k, Value: v})
		}
	}
	for k, v := range c.Extra {
		params = append(params, conf.Param{Name: k, Value: v})
	}
	return params
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkUDP{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
