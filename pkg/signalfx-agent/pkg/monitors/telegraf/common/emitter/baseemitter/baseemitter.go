package baseemitter

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// TelegrafToSFXMetricType returns the signalfx metric type for a telegraf metric
func TelegrafToSFXMetricType(m telegraf.Metric) (datapoint.MetricType, string) {
	switch m.Type() {
	case telegraf.Gauge:
		return datapoint.Gauge, ""
	case telegraf.Counter:
		return datapoint.Counter, ""
	case telegraf.Summary:
		return datapoint.Gauge, "summary"
	case telegraf.Histogram:
		return datapoint.Gauge, "histogram"
	case telegraf.Untyped:
		return datapoint.Gauge, "untyped"
	default:
		return datapoint.Gauge, "unrecognized"
	}
}

// BaseEmitter immediately converts a telegraf measurement into datapoints and
// sends them through Output
type BaseEmitter struct {
	Output              types.Output
	Logger              log.FieldLogger
	OmitPluginDimension bool

	// omittedTags are tags that should be removed from measurements before
	// being processed
	omittedTags map[string]bool
	// addTags are tags that should be added to all measurements
	addTags map[string]string
	// Telegraf has some junk events so we exclude all events by default
	// and can enable them as needed by using IncludeEvent(string) or
	// IncludeEvents([]string).
	// You should look up included metrics using Included(string)bool.
	included map[string]bool
	// excluded metrics and events that should not be emitted.
	// You can add metrics and events to exclude by name using
	// ExcludeDatum(string) and ExcludeData(string).  You should look up
	// excluded events and metrics using Excluded(string)bool
	excluded map[string]bool
	// name map is a map of metric names to their desired metricname
	// this is used for overriding metric names
	nameMap map[string]string
	// metricNameTransformations is an array of functions to apply to parsed metric name
	// from a telegraf metric.
	metricNameTransformations []func(metricName string) string
	// measurementTransformations is an array of functions to apply to an incoming measurement
	// before retrieving the metric name, checking for inclusion/exclusion, etc.
	// Use great discretion with this.
	measurementTransformations []func(telegraf.Metric) error
	datapointTransformations   []func(*datapoint.Datapoint) error
	// whether to omit the "telegraf_type"
	// dimension for documenting original metric type
	omitOriginalMetricType bool
}

// AddTag adds a key/value pair to all measurement tags.  If a key conflicts
// the key value pair in AddTag will override the original key on the
// measurement
func (b *BaseEmitter) AddTag(key string, val string) {
	b.addTags[key] = val
}

// AddTags adds a map of key value pairs to all measurement tags.  If a key
// conflicts the key value pair in AddTags will override the original key on
// the measurement.
func (b *BaseEmitter) AddTags(tags map[string]string) {
	for k, v := range tags {
		b.AddTag(k, v)
	}
}

// IncludeEvent registers an event name to include
// during emission. We disable all events by default because Telegraf has some
// junk events.
func (b *BaseEmitter) IncludeEvent(name string) {
	b.included[name] = true
}

// IncludeEvents registers a list of event names to
// include during emission. We disable all events by default because Telegraf
// has some junk events.
func (b *BaseEmitter) IncludeEvents(names []string) {
	for _, name := range names {
		b.IncludeEvent(name)
	}
}

// Included - checks if events should be included
// during emission.  We disable all events by default because Telegraf has some
// junk events.
func (b *BaseEmitter) Included(name string) bool {
	return b.included[name]
}

// ExcludeDatum adds a name to the list of metrics and events to
// exclude
func (b *BaseEmitter) ExcludeDatum(name string) {
	b.excluded[name] = true
}

// ExcludeData adds a list of names the list of metrics and events
// to exclude
func (b *BaseEmitter) ExcludeData(names []string) {
	for _, name := range names {
		b.ExcludeDatum(name)
	}
}

// IsExcluded - checks if events or metrics should be
// excluded from emission
func (b *BaseEmitter) IsExcluded(name string) bool {
	return b.excluded[name]
}

// OmitTag adds a tag to the list of tags to remove from measurements
func (b *BaseEmitter) OmitTag(tag string) {
	b.omittedTags[tag] = true
}

// OmitTags adds a list of tags the list of tags to remove from measurements
func (b *BaseEmitter) OmitTags(tags []string) {
	for _, tag := range tags {
		b.OmitTag(tag)
	}
}

// FilterTags - filter function for util.CloneAndFilterStringMapWithFunc()
// it returns true if the supplied key is not in the omittedTags map
func (b *BaseEmitter) FilterTags(key string, value string) bool {
	return !b.omittedTags[key]
}

// RenameMetric adds a mapping to rename a metric by it's name
func (b *BaseEmitter) RenameMetric(original string, override string) {
	b.nameMap[original] = override
}

// RenameMetrics takes a map of metric name overrides map[original]override
func (b *BaseEmitter) RenameMetrics(mappings map[string]string) {
	for original, override := range mappings {
		b.RenameMetric(original, override)
	}
}

// GetMetricName parses the metric name and takes name overrides into account
// if a name is overridden it will not have transformations applied to it
func (b *BaseEmitter) GetMetricName(metricName string, field string) string {
	var name string

	if field != "value" {
		name = field
	}

	if name == "" {
		name = metricName
	} else {
		name = fmt.Sprintf("%s.%s", metricName, name)
	}

	if altName := b.nameMap[name]; altName != "" {
		return altName
	}

	// apply metricname transformations
	for _, f := range b.metricNameTransformations {
		name = f(name)
	}

	return name
}

// AddMetricNameTransformation adds a function for mutating metric names.  GetMetricNames()
// will invoke each of the transformation functions after the metric name is parsed
// from the incoming measurement.
func (b *BaseEmitter) AddMetricNameTransformation(f func(string) string) {
	b.metricNameTransformations = append(b.metricNameTransformations, f)
}

// AddMetricNameTransformations adds a list of functions for mutating metric names.  GetMetricNames()
// will invoke each of the transformation functions after the metric name is parsed
// from the incoming measurement.
func (b *BaseEmitter) AddMetricNameTransformations(fns []func(string) string) {
	for _, f := range fns {
		b.AddMetricNameTransformation(f)
	}
}

// AddMeasurementTransformation adds a function to the list of functions the emitter
// will pass an incoming measurement through.  This is useful for manipulating tags
// and fields before the measurement is converted to a SignalFx datapoint.
func (b *BaseEmitter) AddMeasurementTransformation(f func(telegraf.Metric) error) {
	b.measurementTransformations = append(b.measurementTransformations, f)
}

// AddMeasurementTransformations a list of functions to the list of functions the emitter
// will pass an incoming measurement through.  This is useful for manipulating tags
// and fields before the measurement is converted to a SignalFx datapoint.
func (b *BaseEmitter) AddMeasurementTransformations(fns []func(telegraf.Metric) error) {
	for _, f := range fns {
		b.AddMeasurementTransformation(f)
	}
}

// TransformMeasurement applies all measurementTransformations to the supplied measurement
func (b *BaseEmitter) TransformMeasurement(m telegraf.Metric) {
	// apply transformation functions to incoming measurement
	for _, tf := range b.measurementTransformations {
		if err := tf(m); err != nil {
			b.Logger.WithError(err).Errorf("an error occurred applying a transformation to the measurement %v", m)
		}
	}
}

// AddDatapointTransformation adds a callback function that can mutate a
// SignalFx datapoint before it is emitted to the output object.
func (b *BaseEmitter) AddDatapointTransformation(f func(*datapoint.Datapoint) error) {
	b.datapointTransformations = append(b.datapointTransformations, f)
}

func (b *BaseEmitter) AddDatapointTransformations(fns []func(*datapoint.Datapoint) error) {
	for _, f := range fns {
		b.AddDatapointTransformation(f)
	}
}

// TransformMeasurement applies all measurementTransformations to the supplied measurement
func (b *BaseEmitter) TransformDatapoint(dp *datapoint.Datapoint) {
	for _, tf := range b.datapointTransformations {
		if err := tf(dp); err != nil {
			b.Logger.WithError(err).Errorf("An error occurred applying a transformation to the datapoint %v", dp)
		}
	}
}

// AddMetric parses metrics from telegraf and emits them through Output
func (b *BaseEmitter) AddMetric(m telegraf.Metric) {

	// apply transformation functions to the measurement
	b.TransformMeasurement(m)

	// remove tags from metric
	for key := range b.omittedTags {
		m.RemoveTag(key)
	}

	// add tags to metric
	for key, val := range b.addTags {
		m.AddTag(key, val)
	}

	// get metric type and original metric type
	metricType, originalMetricType := TelegrafToSFXMetricType(m)

	// Add common dimensions
	if originalMetricType != "" && !b.omitOriginalMetricType {
		// only add telegraf_type if we override the original type
		m.AddTag("telegraf_type", originalMetricType)
	}

	// add plugin dimension if it doesn't exist
	if !m.HasTag("plugin") && !b.OmitPluginDimension {
		m.AddTag("plugin", m.Name())
	}

	// process fields
	for field, val := range m.Fields() {
		// Generate the metric name
		metricName := b.GetMetricName(m.Name(), field)

		// Check if the metric is explicitly excluded
		if b.IsExcluded(metricName) {
			b.Logger.Debugf("excluding the following metric: %s", metricName)
			continue
		}

		// Get the metric value as a datapoint value
		if metricValue, err := datapoint.CastMetricValue(val); err == nil {
			var dp = datapoint.New(
				metricName,
				m.Tags(),
				metricValue,
				metricType,
				m.Time(),
			)
			b.TransformDatapoint(dp)

			b.Output.SendDatapoints(dp)
		} else {
			// Skip if it's not included
			if !b.Included(metricName) {
				continue
			}
			// We've already type checked field, so set property with value
			metricProps := map[string]interface{}{"message": val}
			var ev = event.NewWithProperties(
				metricName,
				event.AGENT,
				m.Tags(),
				metricProps,
				m.Time(),
			)
			b.Output.SendEvent(ev)
		}
	}
}

// AddError handles errors reported to a telegraf accumulator
func (b *BaseEmitter) AddError(err error) {
	// some telegraf plugins will invoke AddError with nil i.e. sqlserver
	if err != nil {
		b.Logger.WithError(err).Errorf("an error was emitted from the plugin")
	}
}

// AddDebug logs a debug statement
func (b *BaseEmitter) AddDebug(deb string, args ...interface{}) {
	if deb != "" {
		b.Logger.Debugf(deb, args...)
	}
}

// SetOmitOriginalMetricType accepts a boolean to indicate whether the emitter should
// add the original metric type or not to each metric
func (b *BaseEmitter) SetOmitOriginalMetricType(in bool) {
	b.omitOriginalMetricType = in
}

// NewEmitter returns a new BaseEmitter
func NewEmitter(output types.Output, logger log.FieldLogger) *BaseEmitter {
	return &BaseEmitter{
		Output:                     output,
		Logger:                     logger,
		omittedTags:                map[string]bool{},
		included:                   map[string]bool{},
		excluded:                   map[string]bool{},
		addTags:                    map[string]string{},
		nameMap:                    map[string]string{},
		metricNameTransformations:  []func(string) string{},
		measurementTransformations: []func(telegraf.Metric) error{},
	}
}
