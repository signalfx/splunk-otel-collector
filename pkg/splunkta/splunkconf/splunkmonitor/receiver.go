// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkmonitor is a UF-native monitor receiver wrapper. Confmap config
// carries UF stanza fields (include/exclude/charset/truncate/event_breaker +
// resource attrs) and reconstructs the underlying monitorreceiver by delegating
// to tabuilder. This eliminates hand-rolled operator chains.
package splunkmonitor

import (
	"context"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

const TypeStr = "splunk_monitor"

// Config is the UF-native, confmap-serializable monitor config.
// Unknown stanza params are captured in Extra and passed to tabuilder for fidelity.
// Props/Transforms carry the shared props/transforms set, matching splunk_inputs.
type Config struct {
	Include      []string          `mapstructure:"include"`
	Exclude      []string          `mapstructure:"exclude"`
	Index        string            `mapstructure:"index"`
	Source       string            `mapstructure:"source"`
	Sourcetype   string            `mapstructure:"sourcetype"`
	Host         string            `mapstructure:"host"`
	Charset      string            `mapstructure:"charset"`
	Truncate     int               `mapstructure:"truncate"`
	EventBreaker string            `mapstructure:"event_breaker"`
	Props        []conf.Prop       `mapstructure:"props"`
	Transforms   []conf.Transform  `mapstructure:"transforms"`
	Extra        map[string]string `mapstructure:",remain"`
}

type splunkMonitor struct{}

func (splunkMonitor) CreateDefaultConfig() component.Config {
	return &Config{}
}

func (splunkMonitor) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)

	// Reconstruct the stanza name from config (use first include path or a synthetic name)
	stanzaTarget := "/var/log/default"
	if len(c.Include) > 0 {
		stanzaTarget = c.Include[0]
	}
	stanzaName := "monitor://" + stanzaTarget

	params := buildParams(c)
	input := conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   stanzaName,
				Params: params,
			},
		},
	}

	// Delegate to tabuilder.CreateReceiver to build the real monitorreceiver.
	// The shared props/transforms set is fed in, matching splunk_inputs; the
	// monitorreceiver applies per-source matching internally.
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, input, c.Transforms, c.Props, set.TelemetrySettings)
}

func buildParams(c *Config) []conf.Param {
	params := []conf.Param{}

	// Include/exclude paths are stanza metadata, not params, but we'll pass them as params
	// so the underlying monitorreceiver can access them via the stanza.
	// Actually, the target IS in the stanza name; these are additions.
	if len(c.Exclude) > 0 {
		params = append(params, conf.Param{Name: "blacklist", Value: strings.Join(c.Exclude, "|")})
	}

	for k, v := range map[string]string{
		"index":         c.Index,
		"source":        c.Source,
		"sourcetype":    c.Sourcetype,
		"host":          c.Host,
		"CHARSET":       c.Charset,
		"EVENT_BREAKER": c.EventBreaker,
	} {
		if v != "" {
			params = append(params, conf.Param{Name: k, Value: v})
		}
	}
	if c.Truncate > 0 {
		params = append(params, conf.Param{Name: "TRUNCATE", Value: strconv.Itoa(c.Truncate)})
	}

	// Include any extra params from the stanza that weren't explicitly modeled
	for k, v := range c.Extra {
		params = append(params, conf.Param{Name: k, Value: v})
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
