// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package batchreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

func init() {
	_ = featuregate.GlobalRegistry().Set("filelog.allowFileDeletion", true)
}

func NewFactory() receiver.Factory {
	return adapter.NewFactory(batch{
		logger: zap.NewNop(),
	}, component.StabilityLevelAlpha)
}
