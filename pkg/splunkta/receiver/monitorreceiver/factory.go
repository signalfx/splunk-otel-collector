// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package monitorreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// NewFactory creates a new factory for the monitor receiver.
func NewFactory() receiver.Factory {
	return adapter.NewFactory(monitor{
		logger: zap.NewNop(),
	}, component.StabilityLevelAlpha)
}
