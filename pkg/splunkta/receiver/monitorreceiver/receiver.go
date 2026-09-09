// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package monitorreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/move"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/noop"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/operator/prop"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/filter"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/script"
)

type monitor struct {
	logger *zap.Logger
}

// Type is the receiver type
func (monitor) Type() component.Type {
	return component.MustNewType("monitor")
}

// CreateDefaultConfig creates a config with type and version
func (monitor) CreateDefaultConfig() component.Config {
	return createDefaultConfig()
}

func createDefaultConfig() *Config {
	return &Config{}
}

// BaseConfig gets the base config from config
func (monitor) BaseConfig(cfg component.Config) adapter.BaseConfig {
	rcfg := cfg.(Config)
	var operators []operator.Config

	// Insert PCRE whitelist/blacklist filters before any other processing.
	// The log.file.path attribute is set by filelog and available here.
	if w := rcfg.Input.Configuration.Stanza.Params.Get("whitelist"); w != nil && w.Value != "" {
		operators = append(operators, filter.NewWhitelistOperator(w.Value))
	}
	if b := rcfg.Input.Configuration.Stanza.Params.Get("blacklist"); b != nil && b.Value != "" {
		operators = append(operators, filter.NewBlacklistOperator(b.Value))
	}

	operators = append(operators, createSetSourceOperator())

	for _, p := range rcfg.Props {
		ops := prop.CreateOperatorConfigs(p, rcfg.Transforms)
		operators = append(operators, ops...)
	}

	endNoop := noop.NewConfigWithID("end")

	metadata := renameMetadata()
	endNoop.OutputIDs = []string{metadata[0].ID()}
	operators = append(operators, operator.NewConfig(endNoop))
	operators = append(operators, metadata...)

	return adapter.BaseConfig{
		Operators: operators,
	}
}

func createSetSourceOperator() operator.Config {
	c := move.NewConfigWithID("start")
	c.From = entry.NewAttributeField("log.file.path")
	c.To = entry.NewAttributeField("source")
	c.OnError = "send_quiet"
	return operator.NewConfig(c)
}

func (t monitor) InputConfig(config component.Config) operator.Config {
	rcfg := config.(Config)
	oc := file.NewConfig()
	path, err := script.DetermineCommandName(rcfg.BaseDir, rcfg.Input)
	if err != nil {
		t.logger.Error("error reading command", zap.Error(err))
		return operator.NewConfig(oc)
	}
	filter.ApplyIncludeExclude(oc, path, rcfg.Input.Configuration.Stanza, "monitor", t.logger)
	filter.ApplyStanzaConfig(oc, rcfg.Input.Configuration.Stanza)
	return operator.NewConfig(oc)
}

func renameMetadata() []operator.Config {
	source := move.NewConfigWithID("end-source")
	source.From = entry.NewAttributeField("source")
	source.To = entry.NewAttributeField("com.splunk.source")
	source.OnError = "send_quiet"
	source.OutputIDs = []string{"end-sourcetype"}

	sourceType := move.NewConfigWithID("end-sourcetype")
	sourceType.From = entry.NewAttributeField("sourcetype")
	sourceType.To = entry.NewAttributeField("com.splunk.sourcetype")
	sourceType.OnError = "send_quiet"
	sourceType.OutputIDs = []string{"end-host"}

	host := move.NewConfigWithID("end-host")
	host.From = entry.NewAttributeField("host")
	host.To = entry.NewAttributeField("host.name")
	host.OnError = "send_quiet"
	host.OutputIDs = []string{"end-index"}

	index := move.NewConfigWithID("end-index")
	index.From = entry.NewAttributeField("index")
	index.To = entry.NewAttributeField("com.splunk.index")
	index.OnError = "send_quiet"

	return []operator.Config{
		operator.NewConfig(source),
		operator.NewConfig(sourceType),
		operator.NewConfig(host),
		operator.NewConfig(index),
	}
}
