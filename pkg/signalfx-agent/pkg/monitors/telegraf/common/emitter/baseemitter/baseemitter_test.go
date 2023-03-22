package baseemitter

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/neotest"
)

func TestImmediateEmitter_Emit(t *testing.T) {
	type args struct {
		measurement                string
		fields                     map[string]interface{}
		tags                       map[string]string
		metricType                 telegraf.ValueType
		t                          time.Time
		includeEvent               []string
		excludeData                []string
		excludeTag                 []string
		addTag                     map[string]string
		nameMap                    map[string]string
		metricNameTransformations  []func(metricName string) string
		measurementTransformations []func(telegraf.Metric) error
		datapointTransformations   []func(*datapoint.Datapoint) error
	}
	ts := time.Now()
	tests := []struct {
		name           string
		args           args
		wantDatapoints []*datapoint.Datapoint
		wantEvents     []*event.Event
	}{
		{
			name: "emit datapoint without plugin tag",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit datapoint with plugin tag",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType: telegraf.Gauge,
				t:          ts,
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "pluginname",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit datapoint with metric type that defaults to gauge",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType: telegraf.Untyped,
				t:          ts,
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"telegraf_type": "untyped",
						"dim1Key":       "dim1Val",
						"plugin":        "pluginname",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit datapoint and remove an undesirable tag",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				excludeTag: []string{"dim1Key"},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"plugin": "pluginname",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit datapoint and add a tag",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"plugin": "pluginname",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				addTag:     map[string]string{"dim1Key": "dim1Val"},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "pluginname",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit datapoint and add a tag that overrides an original tag",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				addTag:     map[string]string{"dim1Key": "dim1Override"},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Override",
						"plugin":  "pluginname",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "emit an event",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": "hello world",
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				includeEvent: []string{
					"name.fieldname",
				},
			},
			wantEvents: []*event.Event{
				event.NewWithProperties(
					"name.fieldname",
					event.AGENT,
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					map[string]interface{}{
						"message": "hello world",
					},
					ts),
			},
		},
		{
			name: "exclude events that are not explicitly included or sf_metrics",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": "hello world",
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType:   telegraf.Gauge,
				t:            ts,
				includeEvent: []string{},
			},
			wantEvents: nil,
		},
		{
			name: "exclude data by metric name",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
					"plugin":  "pluginname",
				},
				metricType:   telegraf.Gauge,
				t:            ts,
				includeEvent: []string{},
				excludeData:  []string{"name.fieldname"},
			},
			wantEvents:     nil,
			wantDatapoints: nil,
		},
		{
			name: "malformed property objects should be dropped",
			args: args{
				measurement: "objects",
				fields: map[string]interface{}{
					"value": "",
				},
				tags: map[string]string{
					"sf_metric": "objects.host-meta-data",
					"plugin":    "signalfx-metadata",
					"severity":  "4",
				},
				metricType:   telegraf.Gauge,
				t:            ts,
				includeEvent: []string{},
				excludeData:  []string{"name.fieldname"},
			},
			wantEvents:     nil,
			wantDatapoints: nil,
		},
		{
			name: "rename datapoint using metric name map",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				nameMap: map[string]string{
					"name.fieldname": "alt_name.fieldname",
				},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"alt_name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "apply metric name transformation function",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				metricNameTransformations: []func(metricName string) string{
					func(metricName string) string {
						return fmt.Sprintf("transformed.%s", metricName)
					},
				},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"transformed.name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Gauge,
					ts),
			},
		},
		{
			name: "apply datapoint transformation function",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				datapointTransformations: []func(dp *datapoint.Datapoint) error{
					func(dp *datapoint.Datapoint) error {
						dp.MetricType = datapoint.Counter
						return nil
					},
				},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					datapoint.NewIntValue(int64(5)),
					datapoint.Counter,
					ts),
			},
		},
		{
			name: "apply measurement transformation function",
			args: args{
				measurement: "name",
				fields: map[string]interface{}{
					"fieldname": 5,
				},
				tags: map[string]string{
					"dim1Key": "dim1Val",
				},
				metricType: telegraf.Gauge,
				t:          ts,
				measurementTransformations: []func(telegraf.Metric) error{
					func(m telegraf.Metric) error {
						m.AddField("fieldname", 55)
						return nil
					},
				},
			},
			wantDatapoints: []*datapoint.Datapoint{
				datapoint.New(
					"name.fieldname",
					map[string]string{
						"dim1Key": "dim1Val",
						"plugin":  "name",
					},
					datapoint.NewIntValue(int64(55)),
					datapoint.Gauge,
					ts),
			},
		},
	}
	for _, tt := range tests {
		args := tt.args
		wantDatapoints := tt.wantDatapoints
		wantEvents := tt.wantEvents

		t.Run(tt.name, func(t *testing.T) {
			out := neotest.NewTestOutput()
			lg := log.NewEntry(log.New())
			I := NewEmitter(out, lg)
			I.AddTags(args.addTag)
			I.IncludeEvents(args.includeEvent)
			I.ExcludeData(args.excludeData)
			I.OmitTags(args.excludeTag)
			I.RenameMetrics(args.nameMap)
			I.AddMetricNameTransformations(args.metricNameTransformations)
			I.AddMeasurementTransformations(args.measurementTransformations)
			I.AddDatapointTransformations(args.datapointTransformations)
			met, _ := metric.New(args.measurement, args.tags, args.fields, args.t, args.metricType)
			I.AddMetric(met)
			dps := out.FlushDatapoints()
			if !reflect.DeepEqual(dps, wantDatapoints) {
				t.Errorf("actual output: datapoints %v does not match desired: %v", dps, wantDatapoints)
			}

			events := out.FlushEvents()
			if !reflect.DeepEqual(events, wantEvents) {
				t.Errorf("actual output: events %v does not match desired: %v", dps, wantDatapoints)
			}
		})
	}
}

func TestBaseEmitter_AddError(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		var buffer bytes.Buffer
		logger := log.New()
		logger.Out = &buffer
		entry := log.NewEntry(logger)
		err := fmt.Errorf("errorz test")
		B := &BaseEmitter{
			Logger: entry,
		}
		B.AddError(err)
		if !strings.Contains(buffer.String(), "errorz test") {
			t.Errorf("AddError() expected error string with 'errorz test' but got '%s'", buffer.String())
		}
	})
}

func getTelegrafMetricWithOutErr(name string, tags map[string]string, fields map[string]interface{}, t time.Time, metricType telegraf.ValueType) telegraf.Metric {
	m, _ := metric.New(name, tags, fields, t, metricType)
	return m
}

func TestTelegrafToSFXMetricType(t *testing.T) {
	type args struct {
		m telegraf.Metric
	}
	tests := []struct {
		name     string
		args     args
		want     datapoint.MetricType
		wantOrig string
	}{
		{
			name:     "gauge",
			args:     args{getTelegrafMetricWithOutErr("test", map[string]string{}, map[string]interface{}{}, time.Now(), telegraf.Gauge)},
			want:     datapoint.Gauge,
			wantOrig: "",
		},
		{
			name:     "counter",
			args:     args{getTelegrafMetricWithOutErr("test2", map[string]string{}, map[string]interface{}{}, time.Now(), telegraf.Counter)},
			want:     datapoint.Counter,
			wantOrig: "",
		},
		{
			name:     "summary",
			args:     args{getTelegrafMetricWithOutErr("test3", map[string]string{}, map[string]interface{}{}, time.Now(), telegraf.Summary)},
			want:     datapoint.Gauge,
			wantOrig: "summary",
		},
		{
			name:     "histogram",
			args:     args{getTelegrafMetricWithOutErr("test4", map[string]string{}, map[string]interface{}{}, time.Now(), telegraf.Histogram)},
			want:     datapoint.Gauge,
			wantOrig: "histogram",
		},
		{
			name:     "untyped",
			args:     args{getTelegrafMetricWithOutErr("test5", map[string]string{}, map[string]interface{}{}, time.Now(), telegraf.Untyped)},
			want:     datapoint.Gauge,
			wantOrig: "untyped",
		},
	}
	for _, tt := range tests {
		args := tt.args
		want := tt.want
		wantOrig := tt.wantOrig

		t.Run(tt.name, func(t *testing.T) {
			got, got1 := TelegrafToSFXMetricType(args.m)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("TelegrafToSFXMetricType() got = %v, want %v", got, want)
			}
			if got1 != wantOrig {
				t.Errorf("TelegrafToSFXMetricType() gotOrig = %v, want %v", got1, wantOrig)
			}
		})
	}
}
