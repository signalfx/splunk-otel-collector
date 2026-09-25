// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkwineventlog

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const TypeStr = "splunk_wineventlog"

// Config is the UF-native, confmap-serializable wineventlog receiver config.
// Unknown stanza params are captured in Extra and passed to tabuilder for fidelity.
type Config struct {
	EventLogName string            `mapstructure:"event_log_name"`
	Index        string            `mapstructure:"index"`
	Source       string            `mapstructure:"source"`
	Sourcetype   string            `mapstructure:"sourcetype"`
	Host         string            `mapstructure:"host"`
	Props         []conf.Prop       `mapstructure:"props"`
	Transforms    []conf.Transform  `mapstructure:"transforms"`
	Extra        map[string]string `mapstructure:",remain"`
}

type splunkWineventlog struct{}

func (splunkWineventlog) CreateDefaultConfig() component.Config { return &Config{} }

func (splunkWineventlog) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
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
	// Include any extra params from the stanza that weren't explicitly modeled.
	for k, v := range c.Extra {
		params = append(params, conf.Param{Name: k, Value: v})
	}
	input := conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   "wineventlog://" + c.EventLogName,
				Params: params,
			},
		},
	}
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, input, c.Transforms, c.Props, set.TelemetrySettings)
}

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkWineventlog{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
