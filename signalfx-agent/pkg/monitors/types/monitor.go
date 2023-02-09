// Copyright  Splunk, Inc.
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

// Package types exists to avoid circular references between things that need
// to reference common types
package types

// MonitorID is a unique identifier for a specific instance of a monitor
type MonitorID string

// UtilizationMetricPluginName is the name used for the plugin dimension on utilization metrics
const UtilizationMetricPluginName = "signalfx-metadata"
