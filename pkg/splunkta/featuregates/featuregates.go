// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package featuregates defines feature gates for the splunkta package.
package featuregates

import "go.opentelemetry.io/collector/featuregate"

// CookFeatureGate controls whether data is cooked by applying props.conf.
var CookFeatureGate = featuregate.GlobalRegistry().MustRegister(
	"cook",
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, cook the data by applying props.conf"),
	featuregate.WithRegisterFromVersion("v0.1.0"),
)
