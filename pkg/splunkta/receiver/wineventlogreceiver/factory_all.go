// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package wineventlogreceiver

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

// NewFactory creates a new factory for the Windows Event Log receiver (non-Windows stub).
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("wineventlog"),
		func() component.Config {
			return nil
		},
		receiver.WithLogs(
			func(_ context.Context, _ receiver.Settings, _ component.Config, _ consumer.Logs) (receiver.Logs, error) {
				return nil, errors.New("wineventlog is not supported outside Windows environments")
			},
			component.StabilityLevelAlpha,
		),
	)
}
