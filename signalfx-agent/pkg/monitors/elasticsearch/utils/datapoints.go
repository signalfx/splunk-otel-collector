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

package utils

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

func PrepareGaugeHelper(metricName string, dims map[string]string, metricValue *int64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.Gauge(metricName, dims, *metricValue)
}

func PrepareGaugeFHelper(metricName string, dims map[string]string, metricValue *float64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.GaugeF(metricName, dims, *metricValue)
}

func PrepareCumulativeHelper(metricName string, dims map[string]string, metricValue *int64) *datapoint.Datapoint {
	if metricValue == nil {
		return nil
	}
	return sfxclient.Cumulative(metricName, dims, *metricValue)
}
