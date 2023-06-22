// Copyright Splunk, Inc.
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

package telemetry

import (
	"bytes"
	"crypto/md5" // #nosec this is not for cryptographic purposes
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/internal/version"
)

type MetricType string

const (
	DoubleGauge                      MetricType = "DoubleGauge"
	DoubleMonotonicCumulativeSum     MetricType = "DoubleMonotonicCumulativeSum"
	DoubleMonotonicDeltaSum          MetricType = "DoubleMonotonicDeltaSum"
	DoubleMonotonicUnspecifiedSum    MetricType = "DoubleMonotonicUnspecifiedSum"
	DoubleNonmonotonicCumulativeSum  MetricType = "DoubleNonmonotonicCumulativeSum"
	DoubleNonmonotonicDeltaSum       MetricType = "DoubleNonmonotonicDeltaSum"
	DoubleNonmonotonicUnspecifiedSum MetricType = "DoubleNonmonotonicUnspecifiedSum"
	IntGauge                         MetricType = "IntGauge"
	IntMonotonicCumulativeSum        MetricType = "IntMonotonicCumulativeSum"
	IntMonotonicDeltaSum             MetricType = "IntMonotonicDeltaSum"
	IntMonotonicUnspecifiedSum       MetricType = "IntMonotonicUnspecifiedSum"
	IntNonmonotonicCumulativeSum     MetricType = "IntNonmonotonicCumulativeSum"
	IntNonmonotonicDeltaSum          MetricType = "IntNonmonotonicDeltaSum"
	IntNonmonotonicUnspecifiedSum    MetricType = "IntNonmonotonicUnspecifiedSum"
	Summary                          MetricType = "Summary"
	Histogram                        MetricType = "Histogram"
	ExponentialHistogram             MetricType = "ExponentialHistogram"
)

var supportedMetricTypeOptions = fmt.Sprintf(
	"%s, %s, %s, %s, %s, %s,%s, %s, %s, %s, %s, %s, %s, %s",
	DoubleGauge, DoubleMonotonicCumulativeSum,
	DoubleMonotonicDeltaSum, DoubleMonotonicUnspecifiedSum,
	DoubleNonmonotonicCumulativeSum, DoubleNonmonotonicDeltaSum,
	DoubleNonmonotonicUnspecifiedSum, IntGauge,
	IntMonotonicCumulativeSum, IntMonotonicDeltaSum,
	IntMonotonicUnspecifiedSum, IntNonmonotonicCumulativeSum,
	IntNonmonotonicDeltaSum, IntNonmonotonicUnspecifiedSum,
)

var supportedMetricTypes = map[MetricType]bool{
	DoubleGauge: true, DoubleMonotonicCumulativeSum: true,
	DoubleMonotonicDeltaSum: true, DoubleMonotonicUnspecifiedSum: true,
	DoubleNonmonotonicCumulativeSum: true, DoubleNonmonotonicDeltaSum: true,
	DoubleNonmonotonicUnspecifiedSum: true, IntGauge: true,
	IntMonotonicCumulativeSum: true, IntMonotonicDeltaSum: true,
	IntMonotonicUnspecifiedSum: true, IntNonmonotonicCumulativeSum: true,
	IntNonmonotonicDeltaSum: true, IntNonmonotonicUnspecifiedSum: true,
}

// ResourceMetrics is a convenience type for testing helpers and assertions.  Analogous to pdata form, with the exception that
// InstrumentationScope.Metrics items act as both parent metric container and datapoints
// whose identity is based on differing labels and other fields.
type ResourceMetrics struct {
	ResourceMetrics []ResourceMetric `yaml:"resource_metrics"`
}

// ResourceMetric is the top level metric type for a given Resource (set of attributes) and its associated ScopeMetrics.
type ResourceMetric struct {
	Resource     Resource       `yaml:",inline,omitempty"`
	ScopeMetrics []ScopeMetrics `yaml:"scope_metrics"`
}

// ScopeMetrics are the collection of metrics produced by a given InstrumentationScope
type ScopeMetrics struct {
	Scope   InstrumentationScope `yaml:"instrumentation_scope,omitempty"`
	Metrics []Metric             `yaml:"metrics,omitempty"`
}

// Metric is the metric content, representing both the overall definition and a single datapoint.
// TODO: Timestamps
type Metric struct {
	Value       any             `yaml:"value,omitempty"`
	Sum         any             `yaml:"sum,omitempty"`
	Count       any             `yaml:"count,omitempty"`
	Attributes  *map[string]any `yaml:"attributes,omitempty"`
	Name        string          `yaml:"name"`
	Description string          `yaml:"description,omitempty"`
	Unit        string          `yaml:"unit,omitempty"`
	Type        MetricType      `yaml:"type"`
}

// LoadResourceMetrics returns a ResourceMetrics instance generated via parsing a valid yaml file at the provided path.
func LoadResourceMetrics(path string) (*ResourceMetrics, error) {
	metricFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer metricFile.Chdir()

	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(metricFile); err != nil {
		return nil, err
	}
	by := buffer.Bytes()

	var loaded ResourceMetrics
	if err = yaml.UnmarshalStrict(by, &loaded); err != nil {
		return nil, err
	}
	loaded.FillDefaultValues()
	err = loaded.Validate() // in lieu of json/yaml schema adoption
	if err != nil {
		return nil, err
	}
	return &loaded, nil
}

// FillDefaultValues fills ResourceMetrics with default values
func (resourceMetrics *ResourceMetrics) FillDefaultValues() {
	for i, rm := range resourceMetrics.ResourceMetrics {
		rm.Resource.FillDefaultValues()
		for j, sms := range rm.ScopeMetrics {
			if sms.Scope.Version == buildVersionPlaceholder {
				resourceMetrics.ResourceMetrics[i].ScopeMetrics[j].Scope.Version = version.Version
			}

			for _, m := range sms.Metrics {
				if m.Attributes != nil {
					for k, v := range *m.Attributes {
						if v == buildVersionPlaceholder {
							(*m.Attributes)[k] = version.Version
						}
					}
				}
			}
		}
	}
}

// Determines if all values in ResourceMetrics item are valid
func (resourceMetrics ResourceMetrics) Validate() error {
	for _, rm := range resourceMetrics.ResourceMetrics {
		for _, ilm := range rm.ScopeMetrics {
			for _, m := range ilm.Metrics {
				if _, ok := supportedMetricTypes[m.Type]; m.Type != "" && !ok {
					return fmt.Errorf(
						"unsupported MetricType for %s - %s.  Must be one of [%s]",
						m.Name, m.Type, supportedMetricTypeOptions,
					)
				}
			}
		}
	}
	return nil
}

func (resourceMetrics ResourceMetrics) String() string {
	return marshal(resourceMetrics)
}

func (resourceMetric ResourceMetric) String() string {
	return marshal(resourceMetric)
}

func (scopeMetrics ScopeMetrics) String() string {
	return marshal(scopeMetrics)
}

func (metric Metric) String() string {
	return marshal(metric)
}

func (metric Metric) MarshalYAML() (any, error) {
	// fieldalignment causes the Metric yaml rep to be
	// unintuitive so manually unmarshal into map[string]any
	ms := map[string]any{
		"name": metric.Name,
		"type": metric.Type,
	}
	if metric.Unit != "" {
		ms["unit"] = metric.Unit
	}
	if metric.Description != "" {
		ms["description"] = metric.Description
	}
	if metric.Attributes != nil && len(*metric.Attributes) > 0 {
		ms["attributes"] = metric.Attributes
	}
	for _, s := range []struct {
		v any
		k string
	}{
		{k: "sum", v: metric.Sum},
		{k: "count", v: metric.Count},
		{k: "value", v: metric.Value},
	} {
		if s.v != nil {
			ms[s.k] = s.v
		}

	}
	return ms, nil
}

// Provides an md5 hash determined by Metric content.
func (metric Metric) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(metric.String()))) // #nosec
}

// Confirms that all fields, defined or not, in receiver Metric are equal to toCompare.
// TODO: ensure that Metric.Hash equivalence is valid given all possible field values.
func (metric Metric) Equals(toCompare Metric) bool {
	return metric.equals(toCompare, true)
}

// Confirms that all defined fields in receiver Metric are matched in toCompare, ignoring those not set with the
// exception of Labels.  All receiver Metric labels must be equal with those of the candidate to match.
func (metric Metric) RelaxedEquals(toCompare Metric) bool {
	return metric.equals(toCompare, false)
}

// Determines if receiver Metric is equal to toCompare Metric, relaxed if not strict
func (metric Metric) equals(toCompare Metric, strict bool) bool {
	if metric.Name != toCompare.Name && (strict || metric.Name != "") {
		return false
	}
	if metric.Description != toCompare.Description && (strict || metric.Description != "") {
		return false
	}
	if metric.Unit != toCompare.Unit && (strict || metric.Unit != "") {
		return false
	}
	if metric.Type != toCompare.Type && (strict || metric.Type != "") {
		return false
	}

	if metric.Value != toCompare.Value && (strict || metric.Value != nil) {
		return false
	}

	return attributesAreEqual(metric.Attributes, toCompare.Attributes)
}

// FlattenResourceMetrics takes multiple instances of ResourceMetrics and flattens them
// to only unique entries by Resource, InstrumentationScope, and Metric content.
// It will preserve order by removing subsequent occurrences of repeated items
// from the returned flattened ResourceMetrics item
func FlattenResourceMetrics(resourceMetrics ...ResourceMetrics) ResourceMetrics {
	flattened := ResourceMetrics{}

	var resourceHashes []string
	// maps of resource hashes to objects
	resources := map[string]Resource{}
	scopeMetrics := map[string][]ScopeMetrics{}

	// flatten by Resource
	for _, rms := range resourceMetrics {
		for _, rm := range rms.ResourceMetrics {
			resourceHash := rm.Resource.Hash()
			if _, ok := resources[resourceHash]; !ok {
				resources[resourceHash] = rm.Resource
				resourceHashes = append(resourceHashes, resourceHash)
			}
			scopeMetrics[resourceHash] = append(scopeMetrics[resourceHash], rm.ScopeMetrics...)
		}
	}

	// flatten by InstrumentationScope
	for _, resourceHash := range resourceHashes {
		resource := resources[resourceHash]
		resourceMetric := ResourceMetric{
			Resource: resource,
		}

		var ilHashes []string
		// maps of hashes to objects
		ils := map[string]InstrumentationScope{}
		ilMetrics := map[string][]Metric{}
		for _, ilm := range scopeMetrics[resourceHash] {
			ilHash := ilm.Scope.Hash()
			if _, ok := ils[ilHash]; !ok {
				ils[ilHash] = ilm.Scope
				ilHashes = append(ilHashes, ilHash)
			}
			if ilm.Metrics == nil {
				ilm.Metrics = []Metric{}
			}
			ilMetrics[ilHash] = append(ilMetrics[ilHash], ilm.Metrics...)
		}

		// flatten by Metric
		for _, ilHash := range ilHashes {
			il := ils[ilHash]

			var metricHashes []string
			metrics := map[string]Metric{}
			allILMetrics := ilMetrics[ilHash]
			for _, metric := range allILMetrics {
				metricHash := metric.Hash()
				if _, ok := metrics[metricHash]; !ok {
					metrics[metricHash] = metric
					metricHashes = append(metricHashes, metricHash)
				}
			}

			var flattenedMetrics []Metric
			for _, metricHash := range metricHashes {
				flattenedMetrics = append(flattenedMetrics, metrics[metricHash])
			}

			if flattenedMetrics == nil {
				flattenedMetrics = []Metric{}
			}

			sms := ScopeMetrics{
				Scope:   il,
				Metrics: flattenedMetrics,
			}
			resourceMetric.ScopeMetrics = append(resourceMetric.ScopeMetrics, sms)
		}

		flattened.ResourceMetrics = append(flattened.ResourceMetrics, resourceMetric)
	}

	return flattened
}

// ContainsAll determines if everything in `expected` ResourceMetrics is in the receiver ResourceMetrics
// item (i.e. expected ⊆ resourceMetrics). Neither guarantees equivalence, nor that expected contains all of received
// (i.e. is not an expected ≣ resourceMetrics nor resourceMetrics ⊆ expected check).
// Metric equivalence is based on RelaxedEquals() check: fields not in expected (e.g. unit, type, value, etc.)
// are not compared to received, but all labels must match.
// For better reliability, it's advised that both ResourceMetrics items have been flattened by FlattenResourceMetrics.
func (resourceMetrics ResourceMetrics) ContainsAll(expected ResourceMetrics) (bool, error) {
	var missingResources []string
	missingInstrumentationScopes := map[string]struct{}{}
	var missingMetrics []string

	for _, expectedResourceMetric := range expected.ResourceMetrics {
		resourceMatched := false
		for _, resourceMetric := range resourceMetrics.ResourceMetrics {
			if expectedResourceMetric.Resource.Equals(resourceMetric.Resource) {
				resourceMatched = true
				innerMissingInstrumentationScopes := map[string]struct{}{}
				for _, expectedILM := range expectedResourceMetric.ScopeMetrics {
					matchingInstrumentationScope := ""
					for _, ilm := range resourceMetric.ScopeMetrics {
						if expectedILM.Scope.Equals(ilm.Scope) {
							matchingInstrumentationScope = ilm.Scope.String()
							for _, expectedMetric := range expectedILM.Metrics {
								metricFound := false
								for _, metric := range ilm.Metrics {
									if expectedMetric.RelaxedEquals(metric) {
										metricFound = true
									}
								}
								if !metricFound {
									missingMetrics = append(missingMetrics, expectedMetric.String())
								}
							}
							if len(missingMetrics) != 0 {
								return false, fmt.Errorf(
									"%v doesn't contain all of %v. Missing Metrics: %s",
									ilm, expectedILM, missingMetrics)
							}
						}
					}
					if matchingInstrumentationScope != "" {
						// no longer globally missing
						delete(missingInstrumentationScopes, matchingInstrumentationScope)
					} else {
						innerMissingInstrumentationScopes[expectedILM.Scope.String()] = struct{}{}
					}
				}
				if len(innerMissingInstrumentationScopes) != 0 {
					if expectedResourceMetric.Resource.Attributes == nil {
						// since nil attributes will be equal with everything
						// keep track of inner missing scopes globally to be
						// removed above
						for k, v := range innerMissingInstrumentationScopes {
							missingInstrumentationScopes[k] = v
						}
						continue
					}
					var missingIS []string
					for k := range innerMissingInstrumentationScopes {
						missingIS = append(missingIS, k)
					}
					return false, fmt.Errorf(
						"%v doesn't contain all of %v. Missing InstrumentationScopes: %s",
						resourceMetric, expectedResourceMetric, missingIS,
					)
				}
			}
		}
		if !resourceMatched {
			missingResources = append(missingResources, expectedResourceMetric.Resource.String())
		}
	}
	if len(missingInstrumentationScopes) != 0 {
		var missingIS []string
		for k := range missingInstrumentationScopes {
			missingIS = append(missingIS, k)
		}
		return false, fmt.Errorf(
			"%v doesn't contain all of %v. Missing InstrumentationScopes: %s",
			resourceMetrics, expected, missingIS,
		)
	}
	if len(missingResources) != 0 {
		return false, fmt.Errorf(
			"%v doesn't contain all of %v. Missing resources: %s",
			resourceMetrics, expected, missingResources,
		)
	}
	return true, nil
}

// ContainsOnly confirms both resourceMetrics.ContainsAll(expected) and that no additional
// metrics are reported
func (resourceMetrics ResourceMetrics) ContainsOnly(expected ResourceMetrics) (bool, error) {
	if ok, err := resourceMetrics.ContainsAll(expected); !ok {
		return ok, err
	}
	expectedNames := map[string]struct{}{}
	for _, rm := range expected.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				expectedNames[m.Name] = struct{}{}
			}
		}
	}

	unexpectedNames := map[string]struct{}{}
	for _, rm := range resourceMetrics.ResourceMetrics {
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				if _, ok := expectedNames[m.Name]; !ok {
					unexpectedNames[m.Name] = struct{}{}
				}
			}
		}
	}

	if len(unexpectedNames) == 0 {
		return true, nil
	}

	var unexpected []string
	for name := range unexpectedNames {
		unexpected = append(unexpected, name)
	}

	sort.Strings(unexpected)
	return false, fmt.Errorf("%v contains unexpected metrics %s", resourceMetrics, strings.Join(unexpected, ", "))
}
