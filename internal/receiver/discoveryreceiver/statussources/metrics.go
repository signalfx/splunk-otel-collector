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

package statussources

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

// MetricsToReceiverIDs extracts the identifiers receiver creator created receivers will embed in their attributes
func MetricsToReceiverIDs(md pmetric.Metrics) (config.ComponentID, observer.EndpointID) {
	var receiverType, receiverName, endpointID string
	if md.ResourceMetrics().Len() > 0 {
		resourceMetrics := md.ResourceMetrics().At(0)
		mResAttrs := resourceMetrics.Resource().Attributes()
		if r, ok := mResAttrs.Get(discovery.ReceiverTypeAttr); ok {
			receiverType = r.AsString()
		}
		if r, ok := mResAttrs.Get(discovery.ReceiverNameAttr); ok {
			receiverName = r.AsString()
		}
		if r, ok := mResAttrs.Get(discovery.EndpointIDAttr); ok {
			endpointID = r.AsString()
		}
	}
	return config.NewComponentIDWithName(config.Type(receiverType), receiverName), observer.EndpointID(endpointID)
}
