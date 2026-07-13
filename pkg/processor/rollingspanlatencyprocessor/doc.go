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

// Package rollingspanlatencyprocessor contains a processor that maintains a
// rolling EWMA latency baseline per span key and appends a latency.category
// span attribute when a span's duration is statistically anomalous relative to
// its own history.
package rollingspanlatencyprocessor // import "github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor"
