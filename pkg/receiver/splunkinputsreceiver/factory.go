// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

// NewFactory returns a receiver factory for splunk_inputs.
//
// By default, the factory supports tarunner's built-in input stanza schemes.
// Additional schemes, or overrides for built-in schemes, can be registered with
// WithSubReceiver.
func NewFactory(opts ...Option) receiver.Factory {
	options := newFactoryOptions(opts...)
	return receiver.NewFactory(
		component.MustNewType("splunk_inputs"),
		createDefaultConfig,
		receiver.WithLogs(options.createLogsFunc, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return Config{}
}
