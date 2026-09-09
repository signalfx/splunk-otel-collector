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

package components

import (
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/baseline"
	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/lightprometheusreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/promqlreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/signalfxgatewayprometheusremotewritereceiver"
	"github.com/signalfx/splunk-otel-collector/internal/version"
	"github.com/signalfx/splunk-otel-collector/pkg/exporter/splunkoutputsexporter"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension"
	"github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor"
	"github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor"
	"github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/receiver/splunkinputsreceiver"
)

const (
	enableTARunnerFeatureGateID       = "enableTARunner"
	splunkCollectorModule             = "github.com/signalfx/splunk-otel-collector"
	splunkOutputsExporterModule       = splunkCollectorModule + "/pkg/exporter/splunkoutputsexporter"
	oracleEncodingExtensionModule     = splunkCollectorModule + "/pkg/extension/oracleencodingextension"
	smartAgentExtensionModule         = splunkCollectorModule + "/pkg/extension/smartagentextension"
	rollingSpanLatencyProcessorModule = splunkCollectorModule + "/pkg/processor/rollingspanlatencyprocessor"
	timestampProcessorModule          = splunkCollectorModule + "/pkg/processor/timestampprocessor"
	smartAgentReceiverModule          = splunkCollectorModule + "/pkg/receiver/smartagentreceiver"
	splunkInputsReceiverModule        = splunkCollectorModule + "/pkg/receiver/splunkinputsreceiver"
)

var enableTARunner = featuregate.GlobalRegistry().MustRegister(
	enableTARunnerFeatureGateID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the collector supports working with .conf configuration files via the `splunk_inputs` receiver and `splunk_outputs` exporter. "+
		"When disabled (default), the `splunk_inputs` receiver and `splunk_outputs` exporter are not available and the collector will crash if it tries to run them."),
	featuregate.WithRegisterFromVersion("v0.158.0"),
)

// Get returns the public splunk-otel-collector component set: the shared,
// upstream-only baseline plus the Splunk-specific components. The resulting set
// is identical to what the collector shipped before the baseline split.
//
// The public collector is itself a flavor — it layers the Splunk delta below
// onto baseline.NewBaseline(), exactly as a private flavor (appd, UC) would
// layer its own delta. Feature-gated components are appended inline.
func Get() (otelcol.Factories, error) {
	// Every module in this repository is released at version.Version. The
	// explicit versions cover source-built main modules and test binaries whose
	// Go build information omits locally replaced dependencies; release builds
	// still prefer the selected dependency versions recorded by Go.
	b := baseline.NewBaseline(
		baseline.WithModuleVersion(splunkCollectorModule, version.Version),
		baseline.WithModuleVersion(splunkOutputsExporterModule, version.Version),
		baseline.WithModuleVersion(oracleEncodingExtensionModule, version.Version),
		baseline.WithModuleVersion(smartAgentExtensionModule, version.Version),
		baseline.WithModuleVersion(rollingSpanLatencyProcessorModule, version.Version),
		baseline.WithModuleVersion(timestampProcessorModule, version.Version),
		baseline.WithModuleVersion(smartAgentReceiverModule, version.Version),
		baseline.WithModuleVersion(splunkInputsReceiverModule, version.Version),
	)

	b.AddExtensionsWithModulePath(splunkCollectorModule,
		configsourcetelemetryextension.NewFactory(),
		diskqueuestorageextension.NewFactory(),
	)
	b.AddExtensionsWithModulePath(oracleEncodingExtensionModule, oracleencodingextension.NewFactory())
	b.AddExtensionsWithModulePath(smartAgentExtensionModule, smartagentextension.NewFactory())
	b.AddReceiversWithModulePath(splunkCollectorModule,
		discoveryreceiver.NewFactory(),
		gnmireceiver.NewFactory(),
		lightprometheusreceiver.NewFactory(),
		promqlreceiver.NewFactory(),
		signalfxgatewayprometheusremotewritereceiver.NewFactory(),
	)
	b.AddReceiversWithModulePath(smartAgentReceiverModule, smartagentreceiver.NewFactory())
	if enableTARunner.IsEnabled() {
		b.AddReceiversWithModulePath(splunkInputsReceiverModule, splunkinputsreceiver.NewFactory())
		b.AddExportersWithModulePath(splunkOutputsExporterModule, splunkoutputsexporter.NewFactory())
	}
	b.AddProcessorsWithModulePath(timestampProcessorModule, timestampprocessor.NewFactory())
	b.AddProcessorsWithModulePath(rollingSpanLatencyProcessorModule, rollingspanlatencyprocessor.NewFactory())

	return b.Build()
}
