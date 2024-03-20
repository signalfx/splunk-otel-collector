// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
// Temporary copied from the SignalFx exporter to be used in DisableKubeletUtilizationMetrics converter.

package dpfilters // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter/internal/translation/dpfilters"

import "fmt"

type MetricFilter struct {
	Dimensions  map[string]any `mapstructure:"dimensions"`
	MetricName  string         `mapstructure:"metric_name"`
	MetricNames []string       `mapstructure:"metric_names"`
}

func (mf *MetricFilter) normalize() (map[string][]string, error) {
	if mf.MetricName != "" {
		mf.MetricNames = append(mf.MetricNames, mf.MetricName)
	}

	dimSet := map[string][]string{}
	for k, v := range mf.Dimensions {
		switch s := v.(type) {
		case []any:
			var newSet []string
			for _, iv := range s {
				newSet = append(newSet, fmt.Sprintf("%v", iv))
			}
			dimSet[k] = newSet
		case string:
			dimSet[k] = []string{s}
		default:
			return nil, fmt.Errorf("%v should be either a string or string list", v)
		}
	}

	return dimSet, nil
}
