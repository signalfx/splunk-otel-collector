package accumulator

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
)

type testEmitter struct {
	measurement        string
	fields             map[string]interface{}
	tags               map[string]string
	metricType         telegraf.ValueType
	originalMetricType string
	t                  time.Time
	err                error
	deb                string
}

func (e *testEmitter) AddMetric(m telegraf.Metric) {
	e.measurement = m.Name()
	e.fields = m.Fields()
	e.tags = m.Tags()
	e.metricType = m.Type()
	e.t = m.Time()
}
func (e *testEmitter) IncludeEvent(string)       {}
func (e *testEmitter) IncludeEvents([]string)    {}
func (e *testEmitter) ExcludeDatum(string)       {}
func (e *testEmitter) ExcludeData([]string)      {}
func (e *testEmitter) AddTag(string, string)     {}
func (e *testEmitter) AddTags(map[string]string) {}
func (e *testEmitter) OmitTag(string)            {}
func (e *testEmitter) OmitTags([]string)         {}
func (e *testEmitter) AddError(err error) {
	e.err = err
}
func (e *testEmitter) AddDebug(deb string, args ...interface{}) {
	e.deb = fmt.Sprintf(deb, args...)
}

func TestAccumulator(t *testing.T) {
	ac := NewAccumulator(&testEmitter{})
	tests := []struct {
		name string
		want *testEmitter
		fn   func(string, map[string]interface{}, map[string]string, ...time.Time)
	}{
		{
			name: "AddFields()",
			want: &testEmitter{
				measurement:        "field_measurement",
				fields:             map[string]interface{}{"dim1": "dimval1"},
				tags:               map[string]string{"tag1": "tagval1"},
				metricType:         telegraf.Gauge,
				originalMetricType: "untyped",
				t:                  time.Now(),
			},
			fn: ac.AddFields,
		},
		{
			name: "AddGauge()",
			want: &testEmitter{
				measurement:        "gauge_measurement",
				fields:             map[string]interface{}{"dim1": "dimval1"},
				tags:               map[string]string{"tag1": "tagval1"},
				metricType:         telegraf.Gauge,
				originalMetricType: "gauge",
				t:                  time.Now(),
			},
			fn: ac.AddGauge,
		},
		{
			name: "AddCounter()",
			want: &testEmitter{
				measurement:        "counter_measurement",
				fields:             map[string]interface{}{"dim1": "dimval1"},
				tags:               map[string]string{"tag1": "tagval1"},
				metricType:         telegraf.Counter,
				originalMetricType: "counter",
				t:                  time.Now(),
			},
			fn: ac.AddCounter,
		},
		{
			name: "AddSummary()",
			want: &testEmitter{
				measurement:        "summary_measurement",
				fields:             map[string]interface{}{"dim1": "dimval1"},
				tags:               map[string]string{"tag1": "tagval1"},
				metricType:         telegraf.Gauge,
				originalMetricType: "summary",
				t:                  time.Now(),
			},
			fn: ac.AddSummary,
		},
		{
			name: "AddHistogram()",
			want: &testEmitter{
				measurement:        "histogram_measurement",
				fields:             map[string]interface{}{"dim1": "dimval1"},
				tags:               map[string]string{"tag1": "tagval1"},
				metricType:         telegraf.Gauge,
				originalMetricType: "histogram",
				t:                  time.Now(),
			},
			fn: ac.AddHistogram,
		},
	}
	for _, tt := range tests {
		want := tt.want
		fn := tt.fn
		t.Run(tt.name, func(t *testing.T) {
			ac.emit = &testEmitter{}
			fn(want.measurement, want.fields, want.tags, want.t)
			if ac.emit.(*testEmitter).measurement != want.measurement ||
				!reflect.DeepEqual(ac.emit.(*testEmitter).fields, want.fields) ||
				!reflect.DeepEqual(ac.emit.(*testEmitter).tags, want.tags) {
				t.Errorf("Accumulator_AddFields() = %v, want %v", ac.emit, want)
			}
		})
	}
	t.Run("SetPrecision()", func(t *testing.T) {
		ac.emit = &testEmitter{}
		ac.SetPrecision(time.Second*1, time.Second*1)
	})
	t.Run("AddError()", func(t *testing.T) {
		ac.emit = &testEmitter{}
		err := fmt.Errorf("Test Error")
		ac.AddError(err)
		if ac.emit.(*testEmitter).err != err {
			t.Errorf("AddError() = %v, want %v", ac.emit.(*testEmitter).err, err)
		}
	})
}
