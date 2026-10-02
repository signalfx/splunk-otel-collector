// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package splunkscript is a UF-native scripted input receiver wrapper. Its
// confmap config carries the fields of a resolved [script://...] stanza and
// rebuilds the underlying scriptreceiver by delegating to tabuilder.
//
// The type is registered so a .conf config source can emit it. It is
// undocumented and not intended to be written by hand in YAML.
package splunkscript

import (
	"context"
	"errors"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
)

// TypeStr is the UF-native scripted input receiver type.
const TypeStr = "splunk_script"

// Config is the UF-native, confmap-serializable script receiver config. Stanza
// params with no modeled field land in Extra and are passed through to
// tabuilder. Props and Transforms carry the shared props/transforms set,
// matching splunk_inputs.
type Config struct {
	Extra          map[string]string `mapstructure:",remain"`
	ScriptFilename string            `mapstructure:"script_filename"`
	Interval       string            `mapstructure:"interval"`
	Index          string            `mapstructure:"index"`
	Source         string            `mapstructure:"source"`
	Sourcetype     string            `mapstructure:"sourcetype"`
	Host           string            `mapstructure:"host"`
	Props          []conf.Prop       `mapstructure:"props"`
	Transforms     []conf.Transform  `mapstructure:"transforms"`
}

// Validate requires the script path, which is the stanza target and has no
// meaningful default.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.ScriptFilename) == "" {
		return errors.New("splunk_script: script_filename is required")
	}
	return nil
}

type splunkScript struct{}

func (splunkScript) CreateLogs(ctx context.Context, set receiver.Settings, cfg component.Config, nextConsumer consumer.Logs) (receiver.Logs, error) {
	c := cfg.(*Config)
	return tabuilder.CreateReceiver(ctx, "", nextConsumer, c.input(), c.Transforms, c.Props, set.TelemetrySettings)
}

// input rebuilds the resolved stanza this config was emitted from, which is what
// tabuilder dispatches on.
func (c *Config) input() conf.Input {
	params := conf.InputParams(conf.ResourceAttrs{
		Index:      c.Index,
		Source:     c.Source,
		Sourcetype: c.Sourcetype,
		Host:       c.Host,
	}, c.Extra)
	if c.Interval != "" {
		params = append(params, conf.Param{Name: "interval", Value: c.Interval})
	}
	return conf.Input{
		Configuration: conf.Configuration{
			Stanza: conf.Stanza{
				Name:   "script://" + c.ScriptFilename,
				Params: params,
			},
		},
	}
}

// NewFactory returns the UF-native scripted input receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(TypeStr),
		func() component.Config { return &Config{} },
		receiver.WithLogs(splunkScript{}.CreateLogs, component.StabilityLevelAlpha),
	)
}
