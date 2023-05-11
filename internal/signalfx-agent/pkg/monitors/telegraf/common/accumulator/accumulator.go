package accumulator

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter"
)

func getTime(times []time.Time) time.Time {
	if len(times) > 0 {
		return times[0]
	}
	return time.Time{}
}

// Accumulator is an interface used to accumulate telegraf measurements from
// Telegraf plugins.
type Accumulator struct {
	emit emitter.Emitter
}

// AddFields receives a measurement with tags and a time stamp to the accumulator.
// Measurements are passed to the Accumulator's Emitter.
func (ac *Accumulator) AddFields(measurement string, fields map[string]interface{},
	tags map[string]string, t ...time.Time) {
	// as of right now metric always returns a nil error
	m, _ := metric.New(measurement, tags, fields, getTime(t), telegraf.Untyped)
	ac.AddMetric(m)
}

// AddGauge receives a measurement as a "Gauge" with tags and a time stamp to
// the accumulator. Measurements are passed to the Accumulator's Emitter.
func (ac *Accumulator) AddGauge(measurement string, fields map[string]interface{},
	tags map[string]string, t ...time.Time) {
	// as of right now metric always returns a nil error
	m, _ := metric.New(measurement, tags, fields, getTime(t), telegraf.Gauge)
	ac.AddMetric(m)
}

// AddCounter receives a measurement as a "Counter" with tags and a time stamp
// to the accumulator. Measurements are passed to the Accumulator's Emitter.
func (ac *Accumulator) AddCounter(measurement string, fields map[string]interface{},
	tags map[string]string, t ...time.Time) {
	// as of right now metric always returns a nil error
	m, _ := metric.New(measurement, tags, fields, getTime(t), telegraf.Counter)
	ac.AddMetric(m)
}

// AddSummary receives a measurement as a "Counter" with tags and a time stamp
// to the accumulator. Measurements are passed to the Accumulator's Emitter.
func (ac *Accumulator) AddSummary(measurement string, fields map[string]interface{},
	tags map[string]string, t ...time.Time) {
	// as of right now metric always returns a nil error
	m, _ := metric.New(measurement, tags, fields, getTime(t), telegraf.Summary)
	ac.AddMetric(m)
}

// AddHistogram receives a measurement as a "Counter" with tags and a time stamp
// to the accumulator. Measurements are passed to the Accumulator's Emitter.
func (ac *Accumulator) AddHistogram(measurement string, fields map[string]interface{},
	tags map[string]string, t ...time.Time) {
	// as of right now metric always returns a nil error
	m, _ := metric.New(measurement, tags, fields, getTime(t), telegraf.Histogram)
	ac.AddMetric(m)
}

// SetPrecision - SignalFx does not implement this
func (ac *Accumulator) SetPrecision(precision, interval time.Duration) {
}

// AddError - log an error returned by the plugin
func (ac *Accumulator) AddError(err error) {
	ac.emit.AddError(err)
}

// AddMetric - adds a metric and will use the configured
func (ac *Accumulator) AddMetric(m telegraf.Metric) {
	m.Accept() // TODO: mark telegraf tracking metrics correctly in emitter
	ac.emit.AddMetric(m)
}

// AddMetrics - is a convenience function for adding multiple metrics to the
// telegraf accumulator
func (ac *Accumulator) AddMetrics(ms []telegraf.Metric) {
	for _, m := range ms {
		ac.AddMetric(m)
	}
}

// WithTracking - returns an accumulator that supports tracking metrics
func (ac *Accumulator) WithTracking(max int) telegraf.TrackingAccumulator {
	return &TrackingAccumulator{
		Accumulator: ac,
		done:        make(chan telegraf.DeliveryInfo, max),
	}
}

// NewAccumulator returns a pointer to an Accumulator
func NewAccumulator(e emitter.Emitter) *Accumulator {
	return &Accumulator{
		emit: e,
	}
}

// TrackingAccumulator will signal after a metric has been processed.  It indicates
// if the metric was accepted, rejected, or dropped
type TrackingAccumulator struct {
	*Accumulator
	done chan telegraf.DeliveryInfo
}

// AddTrackingMetric - Adds a metric to track on the tracking accumulator
func (ta *TrackingAccumulator) AddTrackingMetric(m telegraf.Metric) telegraf.TrackingID {
	// TODO: maybe we'll care about these at some point,
	// but for now dont worry about tracking metrics
	met, id := metric.WithTracking(m, ta.deliveryCallback)
	ta.AddMetric(met)
	return id
}

// AddTrackingMetricGroup - Adds a group of metrics on the tracking accumulator
func (ta *TrackingAccumulator) AddTrackingMetricGroup(group []telegraf.Metric) telegraf.TrackingID {
	// TODO: maybe we'll care about these at some point,
	// but for now dont worry about tracking metrics
	mets, id := metric.WithGroupTracking(group, ta.deliveryCallback)
	ta.AddMetrics(mets)
	return id
}

// Delivered -
func (ta *TrackingAccumulator) Delivered() <-chan telegraf.DeliveryInfo {
	return ta.done
}

func (ta *TrackingAccumulator) deliveryCallback(i telegraf.DeliveryInfo) {
	select {
	case ta.done <- i:
	default:
		ta.Accumulator.emit.AddDebug("unable to return delivery Status '%t' for ID '%v' channel is full", i.Delivered(), i.ID())
	}
}
