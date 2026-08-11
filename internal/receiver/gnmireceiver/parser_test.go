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

package gnmireceiver

import (
	"testing"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/otel/semconv/v1.22.0"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver/internal/metadata"
)

const testEndpoint = "10.0.0.1:57400"

// countersSubscription matches /interfaces/interface/state/counters with a sum
// default and a couple of leaf overrides.
func countersSubscription() SubscriptionConfig {
	return SubscriptionConfig{
		Path:    "/interfaces/interface/state/counters",
		Mode:    modeSample,
		Default: &MetricConfig{Type: metricTypeSum, Unit: "1"},
		Overrides: map[string]MetricConfig{
			"in-octets":      {Type: metricTypeSum, Unit: "By"},
			"in-octets-rate": {Type: metricTypeGauge, Unit: "By/s"},
			"oper-status":    {Type: metricTypeGauge},
			"enabled":        {Type: metricTypeGauge},
		},
	}
}

func testParser(subs ...SubscriptionConfig) *metricParser {
	if len(subs) == 0 {
		subs = []SubscriptionConfig{countersSubscription()}
	}
	return newMetricParser(testEndpoint, subs)
}

// updateResponse builds a SubscribeResponse for a single leaf under
// /interfaces/interface[name=eth0]/state/counters.
func updateResponse(leaf string, val *gnmipb.TypedValue) *gnmipb.SubscribeResponse {
	return &gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Timestamp: time.Unix(0, 1234).UnixNano(),
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "interfaces"},
						{Name: "interface", Key: map[string]string{"name": "eth0"}},
						{Name: "state"},
						{Name: "counters"},
						{Name: leaf},
					}},
					Val: val,
				}},
			},
		},
	}
}

// onlyMetric asserts exactly one metric was produced and returns it.
func onlyMetric(t *testing.T, m pmetric.Metrics) pmetric.Metric {
	t.Helper()
	require.Equal(t, 1, m.ResourceMetrics().Len())
	sm := m.ResourceMetrics().At(0).ScopeMetrics()
	require.Equal(t, 1, sm.Len())
	require.Equal(t, 1, sm.At(0).Metrics().Len())
	return sm.At(0).Metrics().At(0)
}

func TestParseSyncResponseYieldsNoMetrics(t *testing.T) {
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_SyncResponse{SyncResponse: true},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, m.DataPointCount())
}

func TestParseUintEmitsIntSum(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 12345}}))
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	assert.Equal(t, "interfaces.interface.state.counters.in-octets", metric.Name())
	assert.Equal(t, "By", metric.Unit())
	require.Equal(t, pmetric.MetricTypeSum, metric.Type())
	assert.True(t, metric.Sum().IsMonotonic())
	assert.Equal(t, pmetric.AggregationTemporalityCumulative, metric.Sum().AggregationTemporality())

	dp := metric.Sum().DataPoints().At(0)
	assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
	assert.Equal(t, int64(12345), dp.IntValue())

	name, ok := dp.Attributes().Get("name")
	require.True(t, ok)
	assert.Equal(t, "eth0", name.Str())
}

func TestParsePreservesCounter64Precision(t *testing.T) {
	const large = uint64(1) << 60
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: large}}))
	require.NoError(t, err)

	dp := onlyMetric(t, m).Sum().DataPoints().At(0)
	assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
	assert.Equal(t, int64(large), dp.IntValue())
}

func TestParseIntEmitsSum(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-errors",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_IntVal{IntVal: -7}}))
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	assert.Equal(t, "1", metric.Unit())
	require.Equal(t, pmetric.MetricTypeSum, metric.Type())
	assert.Equal(t, int64(-7), metric.Sum().DataPoints().At(0).IntValue())
}

func TestParseFloatEmitsDoubleGauge(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets-rate",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_FloatVal{FloatVal: 1.5}}))
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	assert.Equal(t, "By/s", metric.Unit())
	require.Equal(t, pmetric.MetricTypeGauge, metric.Type())
	dp := metric.Gauge().DataPoints().At(0)
	assert.Equal(t, pmetric.NumberDataPointValueTypeDouble, dp.ValueType())
	assert.InDelta(t, 1.5, dp.DoubleValue(), 0.0001)
}

func TestParseDoubleEmitsDoubleGauge(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets-rate",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_DoubleVal{DoubleVal: 2.25}}))
	require.NoError(t, err)
	dp := onlyMetric(t, m).Gauge().DataPoints().At(0)
	assert.InDelta(t, 2.25, dp.DoubleValue(), 0.0001)
}

func TestParseBoolEmitsIntGauge(t *testing.T) {
	for _, tt := range []struct {
		name     string
		value    bool
		expected int64
	}{
		{"true", true, 1},
		{"false", false, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m, err := testParser().parse(updateResponse("enabled",
				&gnmipb.TypedValue{Value: &gnmipb.TypedValue_BoolVal{BoolVal: tt.value}}))
			require.NoError(t, err)
			metric := onlyMetric(t, m)
			require.Equal(t, pmetric.MetricTypeGauge, metric.Type())
			assert.Equal(t, tt.expected, metric.Gauge().DataPoints().At(0).IntValue())
		})
	}
}

func TestParseStringEmitsInfoMetric(t *testing.T) {
	m, err := testParser().parse(updateResponse("oper-status",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_StringVal{StringVal: "UP"}}))
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	assert.Equal(t, "interfaces.interface.state.counters.oper-status_info", metric.Name())
	require.Equal(t, pmetric.MetricTypeGauge, metric.Type())
	dp := metric.Gauge().DataPoints().At(0)
	assert.Equal(t, int64(1), dp.IntValue())
	value, ok := dp.Attributes().Get("value")
	require.True(t, ok)
	assert.Equal(t, "UP", value.Str())
}

func TestParseDropsUnconfiguredLeaf(t *testing.T) {
	sub := SubscriptionConfig{
		Path:      "/interfaces/interface/state/counters",
		Mode:      modeSample,
		Overrides: map[string]MetricConfig{"in-octets": {Type: metricTypeSum, Unit: "By"}},
	}
	p := testParser(sub)

	m, err := p.parse(updateResponse("out-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}}))
	require.NoError(t, err)
	assert.Equal(t, 0, m.DataPointCount(), "unconfigured leaf must be dropped")

	m, err = p.parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}}))
	require.NoError(t, err)
	assert.Equal(t, 1, m.DataPointCount())
}

func TestParseDropsUnknownPath(t *testing.T) {
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "system"}, {Name: "memory"}, {Name: "used"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}},
				}},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, m.DataPointCount(), "path outside any subscription must be dropped")
}

func TestParseJoinsNotificationPrefix(t *testing.T) {
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Prefix: &gnmipb.Path{Elem: []*gnmipb.PathElem{
					{Name: "interfaces"},
					{Name: "interface", Key: map[string]string{"name": "eth1"}},
				}},
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "state"}, {Name: "counters"}, {Name: "in-octets"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 5}},
				}},
			},
		},
	})
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	assert.Equal(t, "interfaces.interface.state.counters.in-octets", metric.Name())
	dp := metric.Sum().DataPoints().At(0)
	name, ok := dp.Attributes().Get("name")
	require.True(t, ok)
	assert.Equal(t, "eth1", name.Str())
}

func TestParseOriginPrefixesMetricName(t *testing.T) {
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{
						Origin: "openconfig",
						Elem: []*gnmipb.PathElem{
							{Name: "interfaces"},
							{Name: "interface"},
							{Name: "state"},
							{Name: "counters"},
							{Name: "in-octets"},
						},
					},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}},
				}},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "openconfig.interfaces.interface.state.counters.in-octets",
		onlyMetric(t, m).Name())
}

func TestParseSetsResourceAttributes(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}}))
	require.NoError(t, err)

	res := m.ResourceMetrics().At(0).Resource()
	endpoint, ok := res.Attributes().Get(string(conventions.ServerAddressKey))
	require.True(t, ok)
	assert.Equal(t, testEndpoint, endpoint.Str())
}

func TestParseSetsScopeName(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}}))
	require.NoError(t, err)

	sm := m.ResourceMetrics().At(0).ScopeMetrics().At(0)
	assert.Equal(t, metadata.ScopeName, sm.Scope().Name())
	assert.NotEmpty(t, sm.Scope().Name())
}

func TestParseJSONFlattensObject(t *testing.T) {
	payload := []byte(`{"in-octets": 10, "in-octets-rate": 1.5, "oper-status": "UP"}`)
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "interfaces"},
						{Name: "interface", Key: map[string]string{"name": "eth0"}},
						{Name: "state"},
						{Name: "counters"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_JsonIetfVal{JsonIetfVal: payload}},
				}},
			},
		},
	})
	require.NoError(t, err)

	byName := map[string]pmetric.Metric{}
	ms := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	for i := 0; i < ms.Len(); i++ {
		byName[ms.At(i).Name()] = ms.At(i)
	}
	require.Len(t, byName, 3)

	octets := byName["interfaces.interface.state.counters.in-octets"]
	require.Equal(t, pmetric.MetricTypeSum, octets.Type())
	assert.Equal(t, "By", octets.Unit())
	assert.Equal(t, pmetric.NumberDataPointValueTypeInt, octets.Sum().DataPoints().At(0).ValueType())
	assert.Equal(t, int64(10), octets.Sum().DataPoints().At(0).IntValue())

	rate := byName["interfaces.interface.state.counters.in-octets-rate"]
	require.Equal(t, pmetric.MetricTypeGauge, rate.Type())
	assert.InDelta(t, 1.5, rate.Gauge().DataPoints().At(0).DoubleValue(), 0.0001)

	_, hasInfo := byName["interfaces.interface.state.counters.oper-status_info"]
	assert.True(t, hasInfo, "string leaf should become an info metric")
}

func TestParseJSONNestedObjectExtendsName(t *testing.T) {
	sub := SubscriptionConfig{
		Path:    "/interfaces/interface/state",
		Mode:    modeSample,
		Default: &MetricConfig{Type: metricTypeSum, Unit: "By"},
	}
	payload := []byte(`{"counters": {"in-octets": 4}}`)
	m, err := testParser(sub).parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "interfaces"}, {Name: "interface"}, {Name: "state"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_JsonVal{JsonVal: payload}},
				}},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "interfaces.interface.state.counters.in-octets", onlyMetric(t, m).Name())
}

func TestParseJSONArrayUsesIndexAttribute(t *testing.T) {
	sub := SubscriptionConfig{
		Path:      "/interfaces/interface/state/counters",
		Mode:      modeSample,
		Overrides: map[string]MetricConfig{"queue": {Type: metricTypeSum, Unit: "{packet}"}},
	}
	payload := []byte(`{"queue": [5, 9]}`)
	m, err := testParser(sub).parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "interfaces"},
						{Name: "interface"},
						{Name: "state"},
						{Name: "counters"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_JsonVal{JsonVal: payload}},
				}},
			},
		},
	})
	require.NoError(t, err)

	ms := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	require.Equal(t, 2, ms.Len())
	indexes := map[string]int64{}
	for i := 0; i < ms.Len(); i++ {
		metric := ms.At(i)
		assert.Equal(t, "interfaces.interface.state.counters.queue", metric.Name())
		dp := metric.Sum().DataPoints().At(0)
		idx, ok := dp.Attributes().Get(indexAttr)
		require.True(t, ok, "array element must carry an index attribute")
		indexes[idx.Str()] = dp.IntValue()
	}
	assert.Equal(t, map[string]int64{"0": 5, "1": 9}, indexes)
}

func TestParseInvalidJSONReturnsError(t *testing.T) {
	m, err := testParser().parse(&gnmipb.SubscribeResponse{
		Response: &gnmipb.SubscribeResponse_Update{
			Update: &gnmipb.Notification{
				Update: []*gnmipb.Update{{
					Path: &gnmipb.Path{Elem: []*gnmipb.PathElem{
						{Name: "interfaces"},
						{Name: "interface"},
						{Name: "state"},
						{Name: "counters"},
					}},
					Val: &gnmipb.TypedValue{Value: &gnmipb.TypedValue_JsonVal{JsonVal: []byte(`{"a":`)}},
				}},
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON payload")
	assert.Equal(t, 0, m.DataPointCount())
}

func TestParseLeafListUsesIndexAttribute(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_LeaflistVal{
			LeaflistVal: &gnmipb.ScalarArray{Element: []*gnmipb.TypedValue{
				{Value: &gnmipb.TypedValue_UintVal{UintVal: 11}},
				{Value: &gnmipb.TypedValue_UintVal{UintVal: 22}},
			}},
		}}))
	require.NoError(t, err)

	ms := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	require.Equal(t, 2, ms.Len())
	byIndex := map[string]int64{}
	for i := 0; i < ms.Len(); i++ {
		metric := ms.At(i)
		assert.Equal(t, "interfaces.interface.state.counters.in-octets", metric.Name())
		require.Equal(t, pmetric.MetricTypeSum, metric.Type())
		dp := metric.Sum().DataPoints().At(0)
		assert.Equal(t, pmetric.NumberDataPointValueTypeInt, dp.ValueType())
		idx, ok := dp.Attributes().Get(indexAttr)
		require.True(t, ok, "leaf-list element must carry an index attribute")

		name, ok := dp.Attributes().Get("name")
		require.True(t, ok)
		assert.Equal(t, "eth0", name.Str())
		byIndex[idx.Str()] = dp.IntValue()
	}
	assert.Equal(t, map[string]int64{"0": 11, "1": 22}, byIndex)
}

func TestParseLeafListOfStringsEmitsInfoMetrics(t *testing.T) {
	m, err := testParser().parse(updateResponse("oper-status",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_LeaflistVal{
			LeaflistVal: &gnmipb.ScalarArray{Element: []*gnmipb.TypedValue{
				{Value: &gnmipb.TypedValue_StringVal{StringVal: "UP"}},
				{Value: &gnmipb.TypedValue_StringVal{StringVal: "DOWN"}},
			}},
		}}))
	require.NoError(t, err)

	ms := m.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics()
	require.Equal(t, 2, ms.Len())
	values := map[string]string{}
	for i := 0; i < ms.Len(); i++ {
		metric := ms.At(i)
		assert.Equal(t, "interfaces.interface.state.counters.oper-status_info", metric.Name())
		dp := metric.Gauge().DataPoints().At(0)
		idx, _ := dp.Attributes().Get(indexAttr)
		val, ok := dp.Attributes().Get(infoValueAttr)
		require.True(t, ok)
		values[idx.Str()] = val.Str()
	}
	assert.Equal(t, map[string]string{"0": "UP", "1": "DOWN"}, values)
}

func TestParseEmptyLeafListYieldsNoMetrics(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_LeaflistVal{
			LeaflistVal: &gnmipb.ScalarArray{},
		}}))
	require.NoError(t, err)
	assert.Equal(t, 0, m.DataPointCount())
}

func TestParseUnsupportedValueTypeIsSkipped(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_BytesVal{BytesVal: []byte{1, 2}}}))
	require.NoError(t, err)
	assert.Equal(t, 0, m.DataPointCount())
}

func TestParseUsesLongestMatchingSubscription(t *testing.T) {
	broad := SubscriptionConfig{
		Path:    "/interfaces/interface/state",
		Mode:    modeSample,
		Default: &MetricConfig{Type: metricTypeGauge, Unit: "1"},
	}
	specific := SubscriptionConfig{
		Path:    "/interfaces/interface/state/counters",
		Mode:    modeSample,
		Default: &MetricConfig{Type: metricTypeSum, Unit: "By"},
	}
	m, err := testParser(broad, specific).parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 3}}))
	require.NoError(t, err)

	metric := onlyMetric(t, m)
	require.Equal(t, pmetric.MetricTypeSum, metric.Type())
	assert.Equal(t, "By", metric.Unit())
}

func TestParseUsesNotificationTimestamp(t *testing.T) {
	m, err := testParser().parse(updateResponse("in-octets",
		&gnmipb.TypedValue{Value: &gnmipb.TypedValue_UintVal{UintVal: 1}}))
	require.NoError(t, err)
	dp := onlyMetric(t, m).Sum().DataPoints().At(0)
	assert.Equal(t, int64(1234), dp.Timestamp().AsTime().UnixNano())
}

func TestPathElemNames(t *testing.T) {
	assert.Equal(t, []string{"interfaces", "interface", "state"},
		pathElemNames("/interfaces/interface[name=eth0]/state"))
	assert.Nil(t, pathElemNames("/"))
}
