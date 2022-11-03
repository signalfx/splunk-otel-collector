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

package statussources

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestMetricsToReceiverIDs(t *testing.T) {
	for _, tc := range []struct {
		rType *string
		rName *string
		eID   *string
		name  string
	}{
		{name: "happy path", rType: sPtr("a.type"), rName: sPtr("a.name"), eID: sPtr("an.endpoint")},
		{name: "empty values", rType: sPtr(""), rName: sPtr(""), eID: sPtr("")},
		{name: "missing values", rType: nil, rName: nil, eID: nil},
		{name: "empty receiver type", rType: sPtr(""), rName: sPtr("a.name"), eID: sPtr("an.endpoint")},
		{name: "missing receiver type", rType: nil, rName: sPtr("a.name"), eID: sPtr("an.endpoint")},
		{name: "empty receiver name", rType: sPtr("a.type"), rName: sPtr(""), eID: sPtr("an.endpoint")},
		{name: "missing receiver name", rType: sPtr("a.type"), rName: nil, eID: sPtr("an.endpoint")},
		{name: "empty endpointID", rType: sPtr("a.type"), rName: sPtr("a.name"), eID: sPtr("")},
		{name: "missing endpointID", rType: sPtr("a.type"), rName: sPtr("a.name"), eID: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			md := pmetric.NewMetrics()
			rm := md.ResourceMetrics().AppendEmpty()
			rAttrs := rm.Resource().Attributes()

			if tc.rType != nil {
				rAttrs.PutStr("discovery.receiver.type", *tc.rType)
			}
			if tc.rName != nil {
				rAttrs.PutStr("discovery.receiver.name", *tc.rName)
			}
			if tc.eID != nil {
				rAttrs.PutStr("discovery.endpoint.id", *tc.eID)
			}

			receiverID, endpointID := MetricsToReceiverIDs(md)

			var expectedRType string
			if tc.rType != nil {
				expectedRType = *tc.rType
			}
			require.Equal(t, config.Type(expectedRType), receiverID.Type())

			var expectedRName string
			if tc.rName != nil {
				expectedRName = *tc.rName
			}
			require.Equal(t, expectedRName, receiverID.Name())

			var expectedEndpointID string
			if tc.eID != nil {
				expectedEndpointID = *tc.eID
			}
			require.Equal(t, observer.EndpointID(expectedEndpointID), endpointID)
		})
	}
}

func TestMetricsToReceiverIDsMissingRMetrics(t *testing.T) {
	md := pmetric.NewMetrics()
	receiverID, endpointID := MetricsToReceiverIDs(md)
	require.Equal(t, config.Type(""), receiverID.Type())
	require.Equal(t, "", receiverID.Name())
	require.Equal(t, observer.EndpointID(""), endpointID)
}

func TestMetricsToReceiverIDsMissingRAttributes(t *testing.T) {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().Clear()

	receiverID, endpointID := MetricsToReceiverIDs(md)
	require.Equal(t, config.Type(""), receiverID.Type())
	require.Equal(t, "", receiverID.Name())
	require.Equal(t, observer.EndpointID(""), endpointID)
}

func sPtr(s string) *string {
	return &s
}
