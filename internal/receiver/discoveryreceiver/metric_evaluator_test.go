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

package discoveryreceiver

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestMetricEvaluatorBaseMetricConsumer(t *testing.T) {
	logger := zap.NewNop()
	cfg := &Config{}
	cStore := newCorrelationStore(logger, time.Hour)

	me := newMetricsConsumer(logger, cfg, cStore, nil)
	require.Equal(t, consumer.Capabilities{}, me.Capabilities())

	md := pmetric.NewMetrics()
	require.NoError(t, me.ConsumeMetrics(context.Background(), md))
}

func TestConsumeMetrics(t *testing.T) {
	// If debugging tests, replace the Nop Logger with a test instance to see
	// all statements. Not in regular use to avoid spamming output.
	// logger := zaptest.NewLogger(t)
	logger := zap.NewNop()
	for _, tc := range []struct {
		name  string
		match Match
	}{
		{name: "strict", match: Match{Strict: "desired.name"}},
		{name: "regexp", match: Match{Regexp: "^d[esired]{6}.name$"}},
		{name: "expr", match: Match{Expr: "name == 'desired.name'"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			match := tc.match
			match.Message = "desired body content"
			for _, status := range discovery.StatusTypes {
				match.Status = status
				t.Run(string(status), func(t *testing.T) {
					observerID := component.MustNewIDWithName("an_observer", "observer.name")
					receiverID := component.MustNewIDWithName("a_receiver", "receiver.name")
					cfg := &Config{
						Receivers: map[component.ID]ReceiverEntry{
							receiverID: {
								ServiceType: "a_service",
								Rule:        Rule{text: "a.rule", program: nil},
								Status:      &Status{Metrics: []Match{match}},
							},
						},
						WatchObservers: []component.ID{observerID},
					}
					require.NoError(t, cfg.Validate())

					cStore := newCorrelationStore(logger, time.Hour)

					emitCh := cStore.emitCh
					emitWG := sync.WaitGroup{}
					emitWG.Add(1)
					go func() {
						<-emitCh
						emitWG.Done()
					}()

					endpointID := observer.EndpointID("endpoint.id")
					cStore.UpdateEndpoint(observer.Endpoint{ID: endpointID}, receiverID, observerID)

					ms := &consumertest.MetricsSink{}
					me := newMetricsConsumer(logger, cfg, cStore, ms)

					expectedRes := pcommon.NewResource()
					expectedRes.Attributes().PutStr("discovery.receiver.type", "a_receiver")
					expectedRes.Attributes().PutStr("discovery.receiver.name", "receiver.name")
					expectedRes.Attributes().PutStr("discovery.endpoint.id", "endpoint.id")

					md := pmetric.NewMetrics()

					// This resource should be ignored
					rm := md.ResourceMetrics().AppendEmpty()
					expectedRes.CopyTo(rm.Resource())
					rm.Resource().Attributes().PutStr("extra_attr", "undesired_resource")
					rm.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty().SetName("undesired.name")

					// This resource should be processed
					rm = md.ResourceMetrics().AppendEmpty()
					expectedRes.CopyTo(rm.Resource())
					rm.Resource().Attributes().PutStr("extra_attr", "target_resource")

					sm := rm.ScopeMetrics().AppendEmpty()
					sms := sm.Metrics()
					sms.AppendEmpty().SetName("undesired.name")
					sms.AppendEmpty().SetName("another.undesired.name")
					sms.AppendEmpty().SetName("desired.name")
					sms.AppendEmpty().SetName("desired.name")
					sms.AppendEmpty().SetName("desired.name")

					require.NoError(t, me.ConsumeMetrics(context.Background(), md))

					// wait for the emit channel to be processed
					emitWG.Wait()

					require.Equal(t, map[string]string{
						"service.type":            "a_service",
						"discovery.observer.id":   "an_observer/observer.name",
						"discovery.receiver.name": "receiver.name",
						"discovery.receiver.type": "a_receiver",
						"discovery.status":        string(status),
						"discovery.message":       "desired body content",
						"extra_attr":              "target_resource",
					}, cStore.Attrs(endpointID))

					assert.Equal(t, 1, len(ms.AllMetrics()))
					assert.Equal(t, 2, ms.AllMetrics()[0].ResourceMetrics().Len())
					// Ensure redundant attributes are not added
					for i := 0; i < ms.AllMetrics()[0].ResourceMetrics().Len(); i++ {
						attrs := ms.AllMetrics()[0].ResourceMetrics().At(i).Resource().Attributes()
						_, ok := attrs.Get(discovery.ReceiverTypeAttr)
						assert.False(t, ok)
						_, ok = attrs.Get(discovery.ReceiverNameAttr)
						assert.False(t, ok)
					}
				})
			}
		})
	}
}
