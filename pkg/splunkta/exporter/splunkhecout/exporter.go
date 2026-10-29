// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkhecout is a UF-native HEC output exporter: a thin wrapper over
// the contrib splunkhec exporter whose confmap config carries the resolved
// [hecout] stanza knobs rather than raw splunkhec config.
//
// [hecout] is experimental. It exists to exercise the .conf output path in
// testing, is not a UF output kind splunkd defines, and can be removed at any
// time without notice. The wrapper is where that statement lives, and is why the
// kind is mapped by a registered exporter rather than by the provider itself.
package splunkhecout

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkconf/conf"
)

// TypeStr is the UF-native HEC output exporter type.
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
	params := out.Configuration.Stanza.Params
	param := func(name string) string {
		if p := params.Get(name); p != nil {
			return p.Value
		}
		return ""
	}
	return &Config{
		Endpoint: param("uri"),
		Token:    param("httpEventCollectorToken"),
		TLS:      TLS{InsecureSkipVerify: !params.Bool("sslVerifyServerCert", false)},
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

// NewFactory returns the UF-native HEC output exporter factory.
// It maps its own [hecout] stanzas, so the .conf provider models no output
// kind itself.
func NewFactory() splunkconf.MappingExporterFactory {
	hec := splunkhecexporter.NewFactory()
	return splunkconf.NewMappingExporter(exporter.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		exporter.WithLogs(
			func(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
				hecCfg, err := cfg.(*Config).hecConfig(hec)
				if err != nil {
					return nil, err
				}
				set.Logger.Warn("the [hecout] output stanza is experimental and can be removed at any time without notice")
				// splunkhec rejects settings whose ID type is not its own.
				set.ID = component.NewIDWithName(hec.Type(), set.ID.Name())
				return hec.CreateLogs(ctx, set, hecCfg)
			},
			component.StabilityLevelDevelopment,
		),
	), mapOutputs)
}
