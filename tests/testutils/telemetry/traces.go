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
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// ResourceTraces is a convenience type for test helpers and assertions.
type ResourceTraces struct {
	ResourceSpans []ResourceSpans `yaml:"resource_spans"`
}

// ResourceSpans is the top level grouping of trace data given a Resource (set of attributes) and associated ScopeSpans.
type ResourceSpans struct {
	Resource   Resource     `yaml:",inline,omitempty"`
	ScopeSpans []ScopeSpans `yaml:"scope_spans"`
}

// ScopeSpans is the top level grouping of trace data given InstrumentationScope and associated collection of Span
// instances.
type ScopeSpans struct {
	Scope InstrumentationScope `yaml:"instrumentation_scope,omitempty"`
	Spans []Span               `yaml:"spans,omitempty"`
}

// Span is the trace content, here defined only with the fields required for tests.
type Span struct {
	Attributes *map[string]any `yaml:"attributes,omitempty"`
	Name       string          `yaml:"name,omitempty"`
}

// SaveResourceTraces is a helper function that saves the ResourceTraces to an yaml file.
func (rt *ResourceTraces) SaveResourceTraces(path string) error {
	yamlData, err := yaml.Marshal(rt)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, yamlData, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (rt ResourceTraces) String() string {
	return marshal(rt)
}

// LoadResourceTraces returns a ResourceTraces instance generated via parsing a valid yaml file at the provided path.
func LoadResourceTraces(path string) (*ResourceTraces, error) {
	traceFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer traceFile.Close()

	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(traceFile); err != nil {
		return nil, err
	}
	by := buffer.Bytes()

	var loaded ResourceTraces
	err = yaml.UnmarshalStrict(by, &loaded)
	if err != nil {
		return nil, err
	}

	// Not all spans have data in their instrumentation scope. If a test requires version validation it should include
	// it on the expected data. That's why there is no FillDefaultValues() for ResourceTraces.
	//
	// There is no parsing that requires validation so there isn't a Validate() method either.

	return &loaded, nil
}

func (resourceSpans ResourceSpans) String() string {
	return marshal(resourceSpans)
}

func (scopeSpans ScopeSpans) String() string {
	return marshal(scopeSpans)
}

func (span Span) String() string {
	return marshal(span)
}

func (span Span) RelaxedEquals(toCompare Span) bool {
	return span.Name == toCompare.Name && attributesAreEqual(span.Attributes, toCompare.Attributes)
}

// FlattenResourceTraces takes multiple instances of ResourceTraces and flattens them to only unique entries by
// Resource, InstrumentationScope, and Span contents.
// It will preserve order by removing subsequent occurrences of repeated items from the returned flattened
// ResourceTraces item
func FlattenResourceTraces(resourceTracesSlice ...ResourceTraces) ResourceTraces {
	flattened := ResourceTraces{}

	var resourceHashes []string
	// maps of resource hashes to objects
	resources := map[string]Resource{}
	resourceHashToScopeSpans := map[string][]ScopeSpans{}

	// flatten by Resource
	for _, resourceTraces := range resourceTracesSlice {
		for _, resourceSpans := range resourceTraces.ResourceSpans {
			resourceHash := resourceSpans.Resource.Hash()
			if _, ok := resources[resourceHash]; !ok {
				resources[resourceHash] = resourceSpans.Resource
				resourceHashes = append(resourceHashes, resourceHash)
			}
			resourceHashToScopeSpans[resourceHash] = append(resourceHashToScopeSpans[resourceHash], resourceSpans.ScopeSpans...)
		}
	}

	// flatten by InstrumentationScope
	for _, resourceHash := range resourceHashes {
		resource := resources[resourceHash]
		resourceSpans := ResourceSpans{
			Resource: resource,
		}

		var ilHashes []string
		// maps of hashes to object
		ils := map[string]InstrumentationScope{}
		ilSpans := map[string][]Span{}
		for _, ilss := range resourceHashToScopeSpans[resourceHash] {
			ilHash := ilss.Scope.Hash()
			if _, ok := ils[ilHash]; !ok {
				ils[ilHash] = ilss.Scope
				ilHashes = append(ilHashes, ilHash)
			}
			if ilss.Spans == nil {
				ilss.Spans = []Span{}
			}
			ilSpans[ilHash] = append(ilSpans[ilHash], ilss.Spans...)
		}

		// flatten by Span
		for _, ilHash := range ilHashes {
			il := ils[ilHash]
			allILSpans := ilSpans[ilHash]
			if allILSpans == nil {
				allILSpans = []Span{}
			}
			scopeSpans := ScopeSpans{
				Scope: il,
				Spans: allILSpans,
			}
			resourceSpans.ScopeSpans = append(resourceSpans.ScopeSpans, scopeSpans)
		}

		flattened.ResourceSpans = append(flattened.ResourceSpans, resourceSpans)
	}

	return flattened
}

// ContainsAll determines if everything in `expected` ResourceTraces is in the receiver ResourceTraces
// item (i.e. expected ⊆ resourceTraces). Neither guarantees equivalence, nor that expected contains all of received
// (i.e. is not an expected ≣ resourceTraces nor resourceTraces ⊆ expected check).
// Span equivalence is based on RelaxedEquals() check: fields not in expected (e.g. unit, type, value, etc.)
// are not compared to received, but all labels must match.
// For better reliability, it's advised that both ResourceTraces items have been flattened by FlattenResourceTraces.
func (rt ResourceTraces) ContainsAll(expected ResourceTraces) (bool, error) {
	var missingResources []string
	missingInstrumentationScopes := map[string]struct{}{}
	var missingSpans []string

	for _, expectedResourceSpans := range expected.ResourceSpans {
		resourceMatched := false
		for _, resourceSpans := range rt.ResourceSpans {
			if expectedResourceSpans.Resource.Equals(resourceSpans.Resource) {
				resourceMatched = true
				innerMissingInstrumentationScopes := map[string]struct{}{}
				for _, expectedILS := range expectedResourceSpans.ScopeSpans {
					matchingInstrumentationScope := ""
					for _, ils := range resourceSpans.ScopeSpans {
						if expectedILS.Scope.Equals(ils.Scope) {
							matchingInstrumentationScope = ils.Scope.String()
							for _, expectedSpan := range expectedILS.Spans {
								spanFound := false
								for _, span := range ils.Spans {
									if expectedSpan.RelaxedEquals(span) {
										spanFound = true
									}
								}
								if !spanFound {
									missingSpans = append(missingSpans, expectedSpan.String())
								}
							}
							if len(missingSpans) != 0 {
								return false, fmt.Errorf(
									"%v doesn't contain all of %v. Missing Spans: %s",
									ils.Spans, expectedILS.Spans, missingSpans)
							}
						}
					}
					if matchingInstrumentationScope != "" {
						// no longer globally missing
						delete(missingInstrumentationScopes, matchingInstrumentationScope)
					} else {
						innerMissingInstrumentationScopes[expectedILS.Scope.String()] = struct{}{}
					}
				}
				if len(innerMissingInstrumentationScopes) != 0 {
					if expectedResourceSpans.Resource.Attributes == nil {
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
						resourceSpans.ScopeSpans, expectedResourceSpans.ScopeSpans, missingIS,
					)
				}
			}
		}
		if !resourceMatched {
			missingResources = append(missingResources, expectedResourceSpans.Resource.String())
		}
	}
	if len(missingInstrumentationScopes) != 0 {
		var missingIS []string
		for k := range missingInstrumentationScopes {
			missingIS = append(missingIS, k)
		}
		return false, fmt.Errorf("Missing InstrumentationScopes: %s", missingIS)
	}
	if len(missingResources) != 0 {
		return false, fmt.Errorf(
			"%v doesn't contain all of %v. Missing resources: %s",
			rt.ResourceSpans, expected.ResourceSpans, missingResources,
		)
	}
	return true, nil
}
