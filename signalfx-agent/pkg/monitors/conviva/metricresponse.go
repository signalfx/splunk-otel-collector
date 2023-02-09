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

package conviva

type metricResponse struct {
	Type              string               `json:"type"`
	FilterIDValuesMap map[string][]float64 `json:"filters,omitempty"`
	Meta              *meta                `json:"meta"`
	Tables            map[string]table     `json:"tables,omitempty"`
	Timestamps        []int64              `json:"timestamps,omitempty"`
	Xvalues           []string             `json:"xvalues,omitempty"`
}

type table struct {
	Rows     [][]float64 `json:"rows,omitempty"`
	TotalRow []float64   `json:"total_row,omitempty"`
}

type meta struct {
	Status                int64   `json:"status,omitempty"`
	FiltersWarmup         []int64 `json:"filters_warmup,omitempty"`
	FiltersNotExist       []int64 `json:"filters_not_exist,omitempty"`
	FiltersIncompleteData []int64 `json:"filters_incomplete_data,omitempty"`
}
