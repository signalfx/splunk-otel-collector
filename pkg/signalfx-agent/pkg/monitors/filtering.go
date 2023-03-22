package monitors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
	"github.com/sirupsen/logrus"
)

type monitorFiltering struct {
	filterSet       *dpfilters.FilterSet
	metadata        *Metadata
	hasExtraMetrics bool
}

func newMonitorFiltering(conf config.MonitorCustomConfig, metadata *Metadata) (*monitorFiltering, error) {
	filterSet, err := buildFilterSet(metadata, conf)
	if err != nil {
		return nil, err
	}

	return &monitorFiltering{
		filterSet:       filterSet,
		metadata:        metadata,
		hasExtraMetrics: len(conf.MonitorConfigCore().ExtraMetrics) > 0 || len(conf.MonitorConfigCore().ExtraGroups) > 0,
	}, nil
}

// AddDatapointExclusionFilter to the monitor's filter set.  Make sure you do this
// before any datapoints are sent as it is not thread-safe with SendDatapoint.
func (mf *monitorFiltering) AddDatapointExclusionFilter(filter dpfilters.DatapointFilter) {
	mf.filterSet.ExcludeFilters = append(mf.filterSet.ExcludeFilters, filter)
}

func (mf *monitorFiltering) EnabledMetrics() []string {
	if mf.metadata == nil {
		return nil
	}

	// This can be optimized to cache results on the struct if it is worth it.
	dp := &datapoint.Datapoint{}
	var enabledMetrics []string

	for metric := range mf.metadata.Metrics {
		dp.Metric = metric
		if !mf.filterSet.Matches(dp) {
			enabledMetrics = append(enabledMetrics, metric)
		}
	}

	return enabledMetrics
}

// HasEnabledMetricInGroup returns true if there are any metrics enabled that
// fall into the given group.
func (mf *monitorFiltering) HasEnabledMetricInGroup(group string) bool {
	if mf.metadata == nil {
		return false
	}

	for _, m := range mf.EnabledMetrics() {
		// TODO: If metric names in metadata.yaml ever support wildcards this
		// will have to be enhanced.
		if mf.metadata.Metrics[m].Group == group {
			return true
		}
	}
	return false
}

// HasAnyExtraMetrics returns true if there is any custom metric
// enabled for this output instance.
func (mf *monitorFiltering) HasAnyExtraMetrics() bool {
	return mf.hasExtraMetrics
}

func buildFilterSet(metadata *Metadata, conf config.MonitorCustomConfig) (*dpfilters.FilterSet, error) {
	coreConfig := conf.MonitorConfigCore()

	filter, err := coreConfig.FilterSet()
	if err != nil {
		return nil, err
	}

	excludeFilters := []dpfilters.DatapointFilter{filter}

	// If sendAll is true or there are no metrics don't bother setting up
	// filtering
	if metadata != nil && len(metadata.Metrics) > 0 && !metadata.SendAll {
		// Make a copy of extra metrics from config so we don't alter what the user configured.
		extraMetrics := append([]string{}, coreConfig.ExtraMetrics...)

		// Monitors can add additional extra metrics to allow through such as based on config flags.
		if monitorExtra, ok := conf.(config.ExtraMetrics); ok {
			extraMetrics = append(extraMetrics, monitorExtra.GetExtraMetrics()...)
		}

		includedMetricsFilter, err := newMetricsFilter(metadata, extraMetrics, coreConfig.ExtraGroups)
		if err != nil {
			return nil, fmt.Errorf("unable to construct extraMetrics filter: %s", err)
		}

		// Prepend the included metrics filter.
		excludeFilters = append([]dpfilters.DatapointFilter{dpfilters.Negate(includedMetricsFilter)}, excludeFilters...)
	}

	filterSet := &dpfilters.FilterSet{
		ExcludeFilters: excludeFilters,
	}

	return filterSet, nil
}

var _ dpfilters.DatapointFilter = &extraMetricsFilter{}

// Filter of datapoints based on included status and user configuration of
// extraMetrics and extraGroups.
type extraMetricsFilter struct {
	metadata     *Metadata
	extraMetrics map[string]bool
	stringFilter *filter.BasicStringFilter
}

func validateMetricName(metadata *Metadata, metricName string) error {
	if strings.TrimSpace(metricName) == "" {
		return errors.New("metric name cannot be empty")
	}

	if metadata.SendUnknown {
		// The metrics list isn't exhaustive so can't do extra validation.
		return nil
	}

	if strings.ContainsRune(metricName, '*') {
		// Make sure a metric pattern matches at least one metric.
		m, err := filter.NewBasicStringFilter([]string{metricName})
		if err != nil {
			return err
		}

		for metric := range metadata.Metrics {
			if m.Matches(metric) {
				return nil
			}
		}

		logrus.Warnf("extraMetrics: metric pattern '%s' did not match any available metrics for monitor %s",
			metricName, metadata.MonitorType)
	}

	if !metadata.HasMetric(metricName) {
		logrus.Warnf("extraMetrics: metric '%s' does not exist for monitor %s", metricName, metadata.MonitorType)
	}

	return nil
}

func validateGroup(metadata *Metadata, group string) ([]string, error) {
	if strings.TrimSpace(group) == "" {
		return nil, errors.New("group cannot be empty")
	}

	metrics, ok := metadata.GroupMetricsMap[group]
	if !ok {
		logrus.Warnf("extraMetrics: group %s does not exist for monitor %s", group, metadata.MonitorType)
	}
	return metrics, nil
}

func newMetricsFilter(metadata *Metadata, extraMetrics, extraGroups []string) (*extraMetricsFilter, error) {
	var filterItems []string

	for _, metric := range extraMetrics {
		if err := validateMetricName(metadata, metric); err != nil {
			return nil, err
		}

		// If the user specified a metric that's already included no need to add it.
		if !metadata.DefaultMetrics[metric] {
			filterItems = append(filterItems, metric)
		}
	}

	for _, group := range extraGroups {
		metrics, err := validateGroup(metadata, group)
		if err != nil {
			return nil, err
		}
		filterItems = append(filterItems, metrics...)
	}

	basicFilter, err := filter.NewBasicStringFilter(filterItems)
	if err != nil {
		return nil, fmt.Errorf("unable to construct filter with items %s: %s", filterItems, err)
	}

	effectiveMetrics := map[string]bool{}

	// Precompute set of metrics that matches the filter. This isn't a complete
	// set of metrics that are enabled in the case of metrics that aren't included
	// in metadata. But it provides a fast path for known metrics.
	for metric := range metadata.Metrics {
		if basicFilter.Matches(metric) {
			effectiveMetrics[metric] = true
		}
	}

	return &extraMetricsFilter{metadata, effectiveMetrics, basicFilter}, nil
}

func (mf *extraMetricsFilter) Matches(dp *datapoint.Datapoint) bool {
	if len(mf.metadata.Metrics) == 0 {
		// This monitor has no defined metrics so send everything by default.
		return true
	}

	if !mf.metadata.HasMetric(dp.Metric) && mf.metadata.SendUnknown {
		// This metric is unknown to the metadata and the monitor has requested
		// to send all unknown metrics.
		return true
	}

	if mf.metadata.HasDefaultMetric(dp.Metric) {
		// It's an included metric so send by default.
		return true
	}

	if mf.extraMetrics[dp.Metric] {
		// User has explicitly chosen to send this metric (or a group that this metric belongs to).
		return true
	}

	// Lastly check if it matches filter. If it's a known metric from metadata will get matched
	// above so this is a last check for unknown metrics.
	return mf.stringFilter.Matches(dp.Metric)
}
