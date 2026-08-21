// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package defaults registers every component that ships with this collector distribution.
// Importing it pulls in that entire set of dependencies, including Splunk-internal packages
// resolved only via this module's local replace directives; programs outside this module that
// want to assemble their own, smaller set of components should depend on pkg/components
// instead and never import this package.
package defaults

import (
	"github.com/splunk/tarunner/pkg/splunkinputsreceiver"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/pkg/components"
)

const (
	enableTARunnerFeatureGateID = "enableTARunner"
)

var enableTARunner = featuregate.GlobalRegistry().MustRegister(
	enableTARunnerFeatureGateID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the collector supports working with .conf configuration files via the `splunk_inputs` receiver. "+
		"When disabled (default), the `splunk_inputs` receiver is not available and the collector will crash if it tries to run it."),
	featuregate.WithRegisterFromVersion("v0.158.0"),
)

// Get returns this collector distribution's default set of component factories. It builds
// a fresh components.Registries on every call, so it can be passed directly as
// otelcol.CollectorSettings.Factories. Programs that want to customize the set of
// components, e.g. by calling Register or Deregister, should call components.NewRegistries
// and RegisterDefaults themselves instead of using Get.
func Get() (otelcol.Factories, error) {
	r := components.NewRegistries()
	RegisterDefaults(r)

	if enableTARunner.IsEnabled() {
		r.Receivers.Register(splunkinputsreceiver.NewFactory())
	}

	return r.Build()
}
