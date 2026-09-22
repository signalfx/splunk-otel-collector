// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunktcp is a UF-native TCP receiver wrapper.
// Confmap config carries UF stanza fields (port, index/source/sourcetype/host)
// and reconstructs the underlying tcpreceiver by delegating to tabuilder.
package splunktcp

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

const TypeStr = "splunk_tcp"

// Config is the UF-native, confmap-serializable TCP receiver config.
// Unknown stanza params are captured in Extra and passed to tabuilder for fidelity.
// Props/Transforms carry the shared props/transforms set, matching splunk_inputs.
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

// Validate ensures the config is valid.
func (c *Config) Validate() error {
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535, got %d", c.Port)
	}
	return nil
}

type splunkTCP struct{}

func (splunkTCP) CreateDefaultConfig() component.Config {
	return &Config{}
}

func (splunkTCP) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)

	// Reconstruct the stanza name from config
	target := c.ListenAddress
	if target == "" {
		target = "0.0.0.0"
	}
	if c.Port > 0 {
		target = target + ":" + strconv.Itoa(c.Port)
	}

	stanzaName := "tcp://" + target
	input := conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   stanzaName,
				Params: buildParams(c),
			},
		},
	}

	// Delegate to tabuilder.CreateReceiver to build the real tcpreceiver with props/transforms
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
	// Include any extra params from the stanza that weren't explicitly modeled
	for k, v := range c.Extra {
		params = append(params, conf.Param{Name: k, Value: v})
	}
	return params
}

// NewFactory returns the UF-native TCP receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkTCP{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
