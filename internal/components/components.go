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
	"github.com/signalfx/splunk-otel-collector/pkg/exporter/splunkoutputsexporter"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension"
	"github.com/signalfx/splunk-otel-collector/pkg/processor/timestampprocessor"
	"github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/receiver/splunkinputsreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkbatch"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkmonitor"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkscript"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunktcp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkudp"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/splunkwineventlog"
)

const (
	enableTARunnerFeatureGateID = "enableTARunner"
)

var enableTARunner = featuregate.GlobalRegistry().MustRegister(
	enableTARunnerFeatureGateID,
	featuregate.StageAlpha,
	featuregate.WithRegisterDescription("When enabled, the collector supports working with .conf configuration files via the `splunk_inputs` receiver, the `splunk_outputs` exporter, and the UF-native `splunk_*` receiver types. "+
		"When disabled (default), those components are not available and the collector will crash if it tries to run them."),
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
	b := baseline.NewBaseline()

	b.AddExtensions(
		configsourcetelemetryextension.NewFactory(),
		oracleencodingextension.NewFactory(),
		smartagentextension.NewFactory(),
		diskqueuestorageextension.NewFactory(),
	)
	b.AddReceivers(
		discoveryreceiver.NewFactory(),
		gnmireceiver.NewFactory(),
		lightprometheusreceiver.NewFactory(),
		promqlreceiver.NewFactory(),
		signalfxgatewayprometheusremotewritereceiver.NewFactory(),
		smartagentreceiver.NewFactory(),
	)
	if enableTARunner.IsEnabled() {
		b.AddReceivers(splunkinputsreceiver.NewFactory())
		b.AddExporters(splunkoutputsexporter.NewFactory())

		// UF-native receiver types, one per input stanza scheme. They delegate to
		// the same tabuilder.CreateReceiver that splunk_inputs uses, so the data
		// path is shared. Undocumented: they exist to be emitted from a .conf
		// config source, not written by hand in YAML.
		b.AddReceivers(
			splunkmonitor.NewFactory(),
			splunktcp.NewFactory(),
			splunkudp.NewFactory(),
			splunkscript.NewFactory(),
			splunkbatch.NewFactory(),
			splunkwineventlog.NewFactory(),
		)
	}
	b.AddProcessors(
		timestampprocessor.NewFactory(),
	)

	return b.Build()
}
