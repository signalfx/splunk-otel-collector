// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package wineventlogreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
)

func NewFactory() receiver.Factory {
	return adapter.NewFactory(welreceiver{}, component.StabilityLevelAlpha)
}
