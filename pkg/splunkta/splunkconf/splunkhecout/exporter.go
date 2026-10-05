// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhecout is a UF-native HEC output exporter: a thin wrapper over
// the contrib splunkhec exporter whose confmap config carries the resolved
// [httpout] stanza knobs (uri/token) rather than raw splunkhec config. Like the
// splunk_monitor receiver, it exists so the dotconf provider can emit a
// component TYPE that is not a documented contrib component while reusing the
// real exporter internally.
//
// The exporter side is per-output-kind, symmetric with the per-scheme
// receivers: [httpout] -> HEC maps here; [tcpout] -> S2S is a future
// splunk_s2sout type (not in this prototype). The Go package is unscored
// (splunkhecout); the component type it registers is underscore-separated
// (splunk_hecout).
package splunkhecout

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"
)

// TypeStr is the UF-native HEC output exporter type. Underscore-separated,
// undocumented; no contrib type for a YAML author to write.
const TypeStr = "splunk_hecout"

// Config is the UF-native, confmap-serializable HEC output config. The field
// names match the subset of splunkhec config the provider emits, so the wrapper
// can hand them straight to the contrib exporter.
type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
	TLS      TLS    `mapstructure:"tls"`
}

// TLS is the minimal TLS surface the [httpout] mapping needs.
type TLS struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

// Validate rejects an output with no endpoint or token so a broken outputs.conf
// fails at config time rather than at first export.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("splunk_hecout: endpoint is required")
	}
	if c.Token == "" {
		return fmt.Errorf("splunk_hecout: token is required")
	}
	return nil
}

// hecConfig translates the wrapper Config into a contrib splunkhec Config by
// round-tripping the same keys through confmap. Going through confmap instead of
// reaching into splunkhec struct fields keeps the wrapper robust to upstream
// field renames.
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
	return hecCfg, nil
}

// NewFactory returns the UF-native HEC output exporter factory. It delegates all
// component construction to the contrib splunkhec exporter.
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
				// splunkhec validates that the settings ID type matches its own
				// factory type, so present the delegate's ID under the splunkhec
				// type while keeping this instance's name.
				set.ID = component.NewIDWithName(hec.Type(), set.ID.Name())
				return hec.CreateLogs(ctx, set, hecCfg)
			},
			component.StabilityLevelAlpha,
		),
	)
}
