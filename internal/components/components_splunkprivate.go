// Copyright Splunk, Inc.
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

//go:build splunkprivate

package components

import (
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector-components/exporter/s2sexporter"
	"github.com/signalfx/splunk-otel-collector-components/processor/linebreakprocessor"
	"github.com/signalfx/splunk-otel-collector-components/receiver/wineventlogreceiver"
)

func init() {
	privateExtensionFactories = []extension.Factory{}
	privateReceiverFactories = []receiver.Factory{
		wineventlogreceiver.NewFactory(),
	}
	privateProcessorFactories = []processor.Factory{
		linebreakprocessor.NewFactory(),
	}
	privateExporterFactories = []exporter.Factory{
		s2sexporter.NewFactory(),
	}
	privateConnectorFactories = []connector.Factory{}
}
