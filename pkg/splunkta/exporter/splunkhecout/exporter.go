// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhecout is a UF-native HEC output exporter: a thin wrapper over
// the contrib splunkhec exporter whose confmap config carries the resolved
// [hecout] stanza knobs rather than raw splunkhec config.
//
// It owns the [hecout] translation outright, so the TA runner and a .conf config
// source cannot drift apart. One exporter type per output kind; [tcpout] -> S2S
// is a future splunk_s2sout.
package splunkhecout

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

const TypeStr = "splunk_hecout"

// Config field names match the splunkhec keys they are handed to.
type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
	TLS      TLS    `mapstructure:"tls"`
}

type TLS struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

// ConfigFromOutput maps a resolved [hecout] stanza onto this Config.
func ConfigFromOutput(out conf.Output) *Config {
	param := func(name string) string {
		if p := out.Configuration.Stanza.Params.Get(name); p != nil {
			return p.Value
		}
		return ""
	}
	return &Config{
		Endpoint: param("uri"),
		Token:    param("httpEventCollectorToken"),
		// TODO: wire sslVerifyServerCert from outputs.conf.
		TLS: TLS{InsecureSkipVerify: true},
	}
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("splunk_hecout: endpoint is required")
	}
	if c.Token == "" {
		return errors.New("splunk_hecout: token is required")
	}
	return nil
}

// hecConfig goes through confmap rather than splunkhec's struct fields so an
// upstream field rename does not silently drop a value here.
func (c *Config) hecConfig(f exporter.Factory) (component.Config, error) {
	hecCfg := f.CreateDefaultConfig()
	sub := confmap.NewFromStringMap(map[string]any{
		"endpoint": c.Endpoint,
		"token":    c.Token,
		"tls": map[string]any{
			"insecure_skip_verify": c.TLS.InsecureSkipVerify,
		},
	})
	if err := sub.Unmarshal(hecCfg); err != nil {
		return nil, fmt.Errorf("splunk_hecout: build splunkhec config: %w", err)
	}
	// Core validates this wrapper's Config, not the delegate's, so without this
	// splunkhec's own checks (a malformed endpoint URL among them) never run.
	if v, ok := hecCfg.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("splunk_hecout: invalid splunkhec config: %w", err)
		}
	}
	return hecCfg, nil
}

func NewFactory() exporter.Factory {
	hec := splunkhecexporter.NewFactory()
	return exporter.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		exporter.WithLogs(
			func(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
				hecCfg, err := cfg.(*Config).hecConfig(hec)
				if err != nil {
					return nil, err
				}
				// splunkhec rejects settings whose ID type is not its own.
				set.ID = component.NewIDWithName(hec.Type(), set.ID.Name())
				return hec.CreateLogs(ctx, set, hecCfg)
			},
			component.StabilityLevelAlpha,
		),
	)
}
