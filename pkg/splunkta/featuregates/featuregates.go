// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package featuregates

import "go.opentelemetry.io/collector/featuregate"

var CookFeatureGate = featuregate.GlobalRegistry().MustRegister(
	"cook",
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, cook the data by applying props.conf"),
	featuregate.WithRegisterFromVersion("v0.1.0"),
)
