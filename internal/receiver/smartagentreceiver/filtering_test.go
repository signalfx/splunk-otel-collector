// Copyright 2021 Splunk, Inc.
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

package smartagentreceiver

// Ported from https://github.com/signalfx/signalfx-agent/blob/master/pkg/monitors/filtering_test.go

import (
	"fmt"
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testMetadata(sendUnknown bool) *monitors.Metadata {
	return &monitors.Metadata{
		MonitorType:    "test-monitor",
		DefaultMetrics: utils.StringSet("cpu.idle", "cpu.min", "cpu.max", "mem.used"),
		Metrics: map[string]monitors.MetricInfo{
			"cpu.idle":      {Type: datapoint.Gauge},
			"cpu.min":       {Type: datapoint.Gauge},
			"cpu.max":       {Type: datapoint.Gauge},
			"mem.used":      {Type: datapoint.Counter},
			"mem.free":      {Type: datapoint.Counter},
			"mem.available": {Type: datapoint.Counter}},
		SendUnknown: sendUnknown,
		Groups:      utils.StringSet("cpu", "mem"),
		GroupMetricsMap: map[string][]string{
			// All cpu metrics are included.
			"cpu": {"cpu.idle", "cpu.min", "cpu.max"},
			// Only some mem metrics are included.
			"mem": {"mem.used", "mem.free", "mem.available"},
		},
		SendAll: false,
	}
}

var exhaustiveMetadata = testMetadata(false)
var nonexhaustiveMetadata = testMetadata(true)

func TestDefaultMetrics(t *testing.T) {
	if filter, err := newMetricsFilter(exhaustiveMetadata, nil, nil, zap.NewNop()); err != nil {
		t.Error(err)
	} else {
		// All included metrics should be sent.
		for metric := range exhaustiveMetadata.DefaultMetrics {
			localMetric := metric
			t.Run(fmt.Sprintf("included metric %s should send", metric), func(t *testing.T) {
				dp := &datapoint.Datapoint{
					Metric:     localMetric,
					MetricType: datapoint.Counter,
				}
				if !filter.Matches(dp) {
					t.Error()
				}
			})
		}
	}
}

func TestExtraMetrics(t *testing.T) {
	t.Run("user specifies already-included metric", func(t *testing.T) {
		if filter, err := newMetricsFilter(exhaustiveMetadata, []string{"cpu.idle"}, nil, zap.NewNop()); err != nil {
			t.Error(err)
		} else if filter.extraMetrics["cpu.idle"] {
			t.Error("cpu.idle should not have been in additional metrics because it is already included")
		}
	})

	// Exhaustive
	if filter, err := newMetricsFilter(exhaustiveMetadata, []string{"mem.used"}, nil, zap.NewNop()); err != nil {
		t.Error(err)
	} else {
		for metric, shouldSend := range map[string]bool{
			"mem.used":      true,
			"mem.free":      false,
			"mem.available": false,
		} {
			dp := &datapoint.Datapoint{Metric: metric, MetricType: datapoint.Counter}
			sent := filter.Matches(dp)
			if sent && !shouldSend {
				t.Errorf("metric %s should not have sent", metric)
			}
			if !sent && shouldSend {
				t.Errorf("metric %s should have been sent", metric)
			}
		}
	}

	// Non-exhaustive
	if filter, err := newMetricsFilter(nonexhaustiveMetadata, []string{"dynamic-metric", "some-*"}, nil, zap.NewNop()); err != nil {
		t.Error(err)
	} else {
		for metric, shouldSend := range map[string]bool{
			"dynamic-metric":                  true,
			"some-globbed-metric":             true,
			"unconfigured-and-unknown-metric": true,
			"mem.used":                        true,
			"mem.free":                        false,
		} {
			dp := &datapoint.Datapoint{Metric: metric, MetricType: datapoint.Counter}
			sent := filter.Matches(dp)
			if sent && !shouldSend {
				t.Errorf("metric %s should not have sent", metric)
			}
			if !sent && shouldSend {
				t.Errorf("metric %s should have been sent", metric)
			}
		}
	}
}

func TestGlobbedMetricNames(t *testing.T) {
	if filter, err := newMetricsFilter(exhaustiveMetadata, []string{"mem.*"}, nil, zap.NewNop()); err != nil {
		t.Error(err)
	} else {
		// All memory metrics should be sent.
		metrics := exhaustiveMetadata.GroupMetricsMap["mem"]
		if len(metrics) < 1 {
			t.Fatal("should be checking 1 or more metrics")
		}

		for _, metric := range metrics {
			dp := &datapoint.Datapoint{
				Metric:     metric,
				MetricType: datapoint.Counter,
			}
			if !filter.Matches(dp) {
				t.Errorf("metric %s should have been sent", metric)
			}
		}
	}
}

func TestExtraMetricGroups(t *testing.T) {
	if filter, err := newMetricsFilter(exhaustiveMetadata, nil, []string{"mem"}, zap.NewNop()); err != nil {
		t.Error(err)
	} else {
		for _, metric := range exhaustiveMetadata.GroupMetricsMap["mem"] {
			dp := &datapoint.Datapoint{Metric: metric, MetricType: datapoint.Counter}

			if !filter.Matches(dp) {
				t.Errorf("metric %s should have been sent", metric)
			}
		}
	}
}

func Test_newExtraMetricsFilter(t *testing.T) {
	type args struct {
		metadata     *monitors.Metadata
		extraMetrics []string
		extraGroups  []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// Error cases.
		{"metricName is whitespace", args{exhaustiveMetadata, []string{"    "}, nil}, true},
		{"metricName is invalid regex ", args{exhaustiveMetadata, []string{"^["}, nil}, true},
		{"metricName is invalid regex with wildcard ", args{exhaustiveMetadata, []string{"^[*"}, nil}, true},
		{"groupName is whitespace", args{exhaustiveMetadata, nil, []string{"    "}}, true},
		{"metricName is whitespace", args{nonexhaustiveMetadata, []string{"    "}, nil}, true},
		{"groupName is whitespace", args{nonexhaustiveMetadata, nil, []string{"    "}}, true},

		// Successful cases.
		{"unknown group name", args{exhaustiveMetadata, nil, []string{"unknown-group"}}, false},
		{"no group name or metric name", args{exhaustiveMetadata, nil, nil}, false},
		{"valid group name and unknown metric name", args{exhaustiveMetadata, []string{"unknown-metric"},
			[]string{"cpu"}}, false},
		{"unknown metric name", args{exhaustiveMetadata, []string{"unknown-metric"}, nil}, false},
		{"metric glob doesn't match any metric", args{exhaustiveMetadata, []string{"unknown-metric.*"}, nil}, false},
		{"metric does not exist", args{nonexhaustiveMetadata, []string{"unknown-metric"}, nil}, false},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.name, func(t *testing.T) {
			_, err := newMetricsFilter(tt.args.metadata, tt.args.extraMetrics, tt.args.extraGroups, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("newMetricsFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEmptyMetadata(t *testing.T) {
	filter, err := newMetricsFilter(&monitors.Metadata{}, nil, nil, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, filter)
	dp := &datapoint.Datapoint{Metric: "metric", MetricType: datapoint.Counter}
	require.True(t, filter.Matches(dp))
}

func TestNewMonitorFilteringWithExtraMetrics(t *testing.T) {
	filtering, err := newMonitorFiltering(&prometheusexporter.Config{
		MonitorConfig: config.MonitorConfig{
			Type:                "prometheusexporter",
			DatapointsToExclude: []config.MetricFilter{},
			ExtraMetrics:        []string{"metric"},
		},
		SendAllMetrics: true,
	}, exhaustiveMetadata, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, filtering)

	// 1 ExtraMetrics + 1 metric from SendAllMetrics
	require.Equal(t, 2, len(filtering.filterSet.ExcludeFilters))
}

func TestNewMonitorFilteringInvalid(t *testing.T) {
	tests := []struct {
		err      string
		conf     config.MonitorCustomConfig
		metadata *monitors.Metadata
	}{
		{
			"unable to construct extraMetrics filter: metric name cannot be empty",
			&config.MonitorConfig{
				Type:         "test-monitor",
				ExtraMetrics: []string{"  "},
			},
			exhaustiveMetadata,
		},
		{
			"unable to construct extraMetrics filter: group cannot be empty",
			&config.MonitorConfig{
				Type:        "test-monitor",
				ExtraGroups: []string{"  "},
			},
			exhaustiveMetadata,
		},
		{
			"new filters can't be negated",
			&config.MonitorConfig{
				Type: "test-monitor",
				DatapointsToExclude: []config.MetricFilter{
					{
						MetricName: "metric",
						Negated:    true,
					},
				},
			},
			exhaustiveMetadata,
		},
	}
	for _, test := range tests {
		tt := test

		t.Run(tt.err, func(t *testing.T) {
			filtering, err := newMonitorFiltering(tt.conf, tt.metadata, zap.NewNop())
			require.Nil(t, filtering)
			require.EqualError(t, err, tt.err)
		})
	}
}
