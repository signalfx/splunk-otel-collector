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

package dpfilters

import (
	"errors"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

type overridableDatapointFilter struct {
	dimFilter    filter.StringMapFilter
	metricFilter filter.StringFilter
}

// NewOverridable returns a new overridable filter with the given configuration
func NewOverridable(metricNames []string, dimensions map[string][]string) (DatapointFilter, error) {
	var dimFilter filter.StringMapFilter
	if len(dimensions) > 0 {
		var err error
		dimFilter, err = filter.NewStringMapFilter(dimensions)
		if err != nil {
			return nil, err
		}
	}

	var metricFilter filter.StringFilter
	if len(metricNames) > 0 {
		var err error
		metricFilter, err = filter.NewOverridableStringFilter(metricNames)
		if err != nil {
			return nil, err
		}
	}

	if metricFilter == nil && dimFilter == nil {
		return nil, errors.New("metric filter must have at least one metric or dimension defined on it")
	}

	return &overridableDatapointFilter{
		metricFilter: metricFilter,
		dimFilter:    dimFilter,
	}, nil
}

// Matches tests a datapoint to see whether it is excluded by this filter.
func (f *overridableDatapointFilter) Matches(dp *datapoint.Datapoint) bool {
	return (f.metricFilter == nil || f.metricFilter.Matches(dp.Metric)) &&
		(f.dimFilter == nil || f.dimFilter.Matches(dp.Dimensions))
}
