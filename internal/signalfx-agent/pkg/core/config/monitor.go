package config

import (
	"fmt"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

// MonitorConfig is used to configure monitor instances.  One instance of
// MonitorConfig may be used to configure multiple monitor instances.  If a
// monitor's discovery rule does not match any discovered services, the monitor
// will not run.
type MonitorConfig struct {
	ExtraDimensionsFromEndpoint map[string]string      `yaml:"extraDimensionsFromEndpoint" json:"extraDimensionsFromEndpoint"`
	ExtraSpanTags               map[string]string      `yaml:"extraSpanTags" json:"extraSpanTags"`
	OtherConfig                 map[string]interface{} `yaml:",inline" neverLog:"omit"`
	ExtraDimensions             map[string]string      `yaml:"extraDimensions" json:"extraDimensions"`
	ConfigEndpointMappings      map[string]string      `yaml:"configEndpointMappings" json:"configEndpointMappings"`
	ExtraSpanTagsFromEndpoint   map[string]string      `yaml:"extraSpanTagsFromEndpoint" json:"extraSpanTagsFromEndpoint"`
	DefaultSpanTags             map[string]string      `yaml:"defaultSpanTags" json:"defaultSpanTags"`
	DimensionTransformations    map[string]string      `yaml:"dimensionTransformations" json:"dimensionTransformations"`
	ValidateDiscoveryRule       *bool                  `yaml:"validateDiscoveryRule"`
	DefaultSpanTagsFromEndpoint map[string]string      `yaml:"defaultSpanTagsFromEndpoint" json:"defaultSpanTagsFromEndpoint"`
	DiscoveryRule               string                 `yaml:"discoveryRule" json:"discoveryRule"`
	MonitorID                   types.MonitorID        `yaml:"-" hash:"ignore"`
	Type                        string                 `yaml:"type" json:"type"`
	ValidationError             string                 `yaml:"-" json:"-" hash:"ignore"`
	ProcPath                    string                 `yaml:"-" json:"-"`
	Hostname                    string                 `yaml:"-" json:"-"`
	BundleDir                   string                 `yaml:"-" json:"-"`
	DatapointsToExclude         []MetricFilter         `yaml:"datapointsToExclude" json:"datapointsToExclude" default:"[]"`
	ExtraGroups                 []string               `yaml:"extraGroups" json:"extraGroups"`
	ExtraMetrics                []string               `yaml:"extraMetrics" json:"extraMetrics"`
	MetricNameTransformations   yaml.MapSlice          `yaml:"metricNameTransformations"`
	IntervalSeconds             int                    `yaml:"intervalSeconds" json:"intervalSeconds"`
	IsolatedCollectd            bool                   `yaml:"isolatedCollectd" json:"isolatedCollectd"`
	DisableEndpointDimensions   bool                   `yaml:"disableEndpointDimensions" json:"disableEndpointDimensions"`
	DisableHostDimensions       bool                   `yaml:"disableHostDimensions" json:"disableHostDimensions" default:"false"`
	Solo                        bool                   `yaml:"solo" json:"solo"`
}

// Validate ensures the config is correct beyond what basic YAML parsing
// ensures
func (mc *MonitorConfig) Validate() error {
	var err error
	if _, err = mc.FilterSet(); err != nil {
		return err
	}
	if _, err = mc.MetricNameExprs(); err != nil {
		return err
	}
	return nil
}

// FilterSet makes a filter set using the new filter style
func (mc *MonitorConfig) FilterSet() (*dpfilters.FilterSet, error) {
	return makeNewFilterSet(mc.DatapointsToExclude)
}

type RegexpWithReplace struct {
	Regexp      *regexp.Regexp
	Replacement string
}

func (mc *MonitorConfig) MetricNameExprs() ([]*RegexpWithReplace, error) {
	var out []*RegexpWithReplace
	for _, pair := range mc.MetricNameTransformations {
		k, ok := pair.Key.(string)
		if !ok {
			return nil, fmt.Errorf("metricNameTransformation key not a string")
		}

		v, ok := pair.Value.(string)
		if !ok {
			return nil, fmt.Errorf("metricNameTransformation value not a string")
		}

		re, err := regexp.Compile("^" + k + "$")
		if err != nil {
			return nil, err
		}
		out = append(out, &RegexpWithReplace{re, v})
	}
	return out, nil
}

// MonitorConfigCore provides a way of getting the MonitorConfig when embedded
// in a struct that is referenced through a more generic interface.
func (mc *MonitorConfig) MonitorConfigCore() *MonitorConfig {
	return mc
}

// IsCollectdBased returns whether this monitor type depends on the
// collectd subprocess to run.
func (mc *MonitorConfig) IsCollectdBased() bool {
	return strings.HasPrefix(mc.Type, "collectd/")
}

// MonitorCustomConfig represents monitor-specific configuration that doesn't
// appear in the MonitorConfig struct.
type MonitorCustomConfig interface {
	MonitorConfigCore() *MonitorConfig
}

// ExtraMetrics interface for monitors that support generating additional metrics to allow through.
type ExtraMetrics interface {
	GetExtraMetrics() []string
}
