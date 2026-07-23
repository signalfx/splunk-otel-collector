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

package components

import (
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
)

// additionalExtensions, additionalReceivers, additionalExporters, additionalProcessors, and
// additionalConnectors are merged into their respective factory maps in Get().
//
// They are empty in normal builds. A CI-generated, uncommitted file
// (components_private_gen.go, produced by .gitlab/build-private-components.sh) can populate
// these slices from an init() function to register additional components at build time,
// without editing this repository. See .gitlab/build-private-components.sh for details.
var (
	additionalExtensions []extension.Factory
	additionalReceivers  []receiver.Factory
	additionalExporters  []exporter.Factory
	additionalProcessors []processor.Factory
	additionalConnectors []connector.Factory
)
