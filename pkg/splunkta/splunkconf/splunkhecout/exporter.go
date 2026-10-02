// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhecout is a UF-native HEC output exporter: a thin wrapper over
// the contrib splunkhec exporter whose confmap config carries the resolved
// [hecout] stanza knobs (uri/token) rather than raw splunkhec config. Like the
// splunk_monitor receiver, it exists so a .conf config source can emit a
// component type that is not a documented contrib component, while reusing the
// real exporter internally.
//
// This package is the single owner of the [hecout] translation. ConfigFromOutput
// maps the stanza params, and the factory maps the resulting Config onto
// splunkhec, so the TA runner and any .conf config source agree by construction.
//
// The exporter side is per-output-kind, symmetric with the per-scheme receivers:
// [hecout] -> HEC maps here; [tcpout] -> S2S is a future splunk_s2sout type. The
// Go package name is unscored (splunkhecout); the component type it registers is
// underscore-separated (splunk_hecout).
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

// TypeStr is the UF-native HEC output exporter type. Underscore-separated,
// undocumented; no contrib type for a YAML author to write.
const TypeStr = "splunk_hecout"

// Config is the UF-native, confmap-serializable HEC output config. The field
// names match the subset of splunkhec config this wrapper emits, so the values
// can be handed straight to the contrib exporter.
type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
	TLS      TLS    `mapstructure:"tls"`
}

// TLS is the minimal TLS surface the [hecout] mapping needs.
type TLS struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

// ConfigFromOutput maps a resolved [hecout] stanza onto this Config. It is the
// one place the .conf key names are translated, so the TA runner and a .conf
// config source cannot drift apart.
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

// Validate rejects an output with no endpoint or token so a broken outputs.conf
// fails at config time rather than at first export.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("splunk_hecout: endpoint is required")
	}
	if c.Token == "" {
		return errors.New("splunk_hecout: token is required")
	}
	return nil
}

// hecConfig translates this Config into a contrib splunkhec Config by
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
	// splunkhec checks things this wrapper does not model, a malformed endpoint
	// URL among them. Its Validate does not run on its own here because the
	// component core validates is this wrapper's Config, not the delegate's.
	if v, ok := hecCfg.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("splunk_hecout: invalid splunkhec config: %w", err)
		}
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
