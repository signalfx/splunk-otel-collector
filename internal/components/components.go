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
	"go.opentelemetry.io/collector/otelcol"

	"github.com/signalfx/splunk-otel-collector/baseline"
	"github.com/signalfx/splunk-otel-collector/internal/exporter/splunkmodinputexporter"
	"github.com/signalfx/splunk-otel-collector/internal/extension/configsourcetelemetryextension"
	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension"
	"github.com/signalfx/splunk-otel-collector/internal/extension/splunkauthextension"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/promqlreceiver"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/ptpreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/oracleencodingextension"
	"github.com/signalfx/splunk-otel-collector/pkg/extension/smartagentextension"
	"github.com/signalfx/splunk-otel-collector/pkg/receiver/smartagentreceiver"
)

// Get returns the public splunk-otel-collector component set: the shared,
// upstream-only baseline plus the Splunk-specific components. The resulting set
// is identical to what the collector shipped before the baseline split.
//
// The public collector is itself a flavor — it layers the Splunk delta below
// onto baseline.NewBaseline(), exactly as a private flavor (appd, UC) would
// layer its own delta.
func Get() (otelcol.Factories, error) {
	b := baseline.NewBaseline()

	b.AddExporters(splunkmodinputexporter.NewFactory())
	b.AddExtensions(
		splunkauthextension.NewFactory(),
		configsourcetelemetryextension.NewFactory(),
		oracleencodingextension.NewFactory(),
		smartagentextension.NewFactory(),
		diskqueuestorageextension.NewFactory(),
	)
	b.AddReceivers(
		discoveryreceiver.NewFactory(),
		gnmireceiver.NewFactory(),
		promqlreceiver.NewFactory(),
		ptpreceiver.NewFactory(),
		smartagentreceiver.NewFactory(),
	)

	return b.Build()
}
