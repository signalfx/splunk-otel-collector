// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package udpreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/udp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/move"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/transformer/noop"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/operator/prop"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.opentelemetry.io/collector/component"
)

type monitor struct{}

// Type is the receiver type
func (monitor) Type() component.Type {
	return component.MustNewType("udp")
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

func (t monitor) InputConfig(config component.Config) operator.Config {
	rcfg := config.(Config)
	oc := udp.NewConfig()
	listenAddress, _ := stanza.ParseName(rcfg.Input.Configuration.Stanza.Name)
	oc.ListenAddress = stanza.ListenAddress(listenAddress.Target)
	oc.Attributes = map[string]helper.ExprStringConfig{}
	if hostParam := rcfg.Input.Configuration.Stanza.Params.Get("host"); hostParam != nil {
		// TODO: find a way to run host detection when requested.
		oc.Attributes["host"] = helper.ExprStringConfig(hostParam.Value)
	}

	if indexParam := rcfg.Input.Configuration.Stanza.Params.Get("index"); indexParam != nil {
		oc.Attributes["index"] = helper.ExprStringConfig(indexParam.Value)
	}

	if sourceTypeParam := rcfg.Input.Configuration.Stanza.Params.Get("sourcetype"); sourceTypeParam != nil {
		oc.Attributes["sourcetype"] = helper.ExprStringConfig(sourceTypeParam.Value)
	}

	if sourceParam := rcfg.Input.Configuration.Stanza.Params.Get("source"); sourceParam != nil {
		oc.Attributes["source"] = helper.ExprStringConfig(sourceParam.Value)
	}

	oc.SplitConfig.LineEndPattern = ""
	oc.SplitConfig.OmitPattern = false
	oc.SplitConfig.LineStartPattern = ""
	oc.Encoding = "nop"

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
