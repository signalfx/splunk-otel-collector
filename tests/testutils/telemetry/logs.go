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
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/tests/internal/version"
)

// ResourceLogs is a convenience type for testing helpers and assertions.
// Analogous to pdata form, with the exception that InstrumentationScope.Logs items act as both parent log container
// and records whose identity is based on differing attributes and other fields.
type ResourceLogs struct {
	ResourceLogs []ResourceLog `yaml:"resource_logs"`
}

// ResourceLog is the top level log type for a given Resource (set of attributes) and its associated ScopeLogs.
type ResourceLog struct {
	Resource  Resource    `yaml:",inline,omitempty"`
	ScopeLogs []ScopeLogs `yaml:"scope_logs"`
}

// ScopeLogs are the collection of logs produced by a given InstrumentationScope
type ScopeLogs struct {
	Scope InstrumentationScope `yaml:"instrumentation_scope,omitempty"`
	Logs  []Log                `yaml:"logs,omitempty"`
}

// Log is the log content, representing both the overall definition and a single datapoint.
type Log struct {
	ObservedTimestamp time.Time            `yaml:"observed_timestamp,omitempty"`
	Timestamp         time.Time            `yaml:"timestamp,omitempty"`
	Body              any                  `yaml:"body,omitempty"`
	Attributes        *map[string]any      `yaml:"attributes,omitempty"`
	Severity          *plog.SeverityNumber `yaml:"severity,omitempty"`
	SeverityText      string               `yaml:"severity_text,omitempty"`
}

// LoadResourceLogs returns a ResourceLogs instance generated via parsing a valid yaml file at the provided path.
func LoadResourceLogs(path string) (*ResourceLogs, error) {
	logFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer logFile.Chdir()

	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(logFile); err != nil {
		return nil, err
	}
	by := buffer.Bytes()

	var loaded ResourceLogs
	err = yaml.UnmarshalStrict(by, &loaded)
	if err != nil {
		return nil, err
	}
	loaded.FillDefaultValues()
	err = loaded.Validate() // in lieu of json/yaml schema adoption
	if err != nil {
		return nil, err
	}
	return &loaded, nil
}

// FillDefaultValues fills ResourceLogs with default values
func (resourceLogs *ResourceLogs) FillDefaultValues() {
	for i, rl := range resourceLogs.ResourceLogs {
		rl.Resource.FillDefaultValues()
		for j, sls := range rl.ScopeLogs {
			if sls.Scope.Version == buildVersionPlaceholder {
				resourceLogs.ResourceLogs[i].ScopeLogs[j].Scope.Version = version.Version
			}
			for _, sl := range sls.Logs {
				if sl.Attributes != nil {
					for k, v := range *sl.Attributes {
						if v == buildVersionPlaceholder {
							(*sl.Attributes)[k] = version.Version
						}
					}
				}
			}
		}
	}
}

// Determines if all values in ResourceLogs item are valid
func (resourceLogs ResourceLogs) Validate() error {
	for _, rm := range resourceLogs.ResourceLogs {
		for _, sls := range rm.ScopeLogs {
			for range sls.Logs {
				continue
			}
		}
	}
	return nil
}

func (resourceLogs ResourceLogs) String() string {
	return marshal(resourceLogs)
}

func (resourceLog ResourceLog) String() string {
	return marshal(resourceLog)
}

func (scopeLogs ScopeLogs) String() string {
	return marshal(scopeLogs)
}

func (log Log) String() string {
	return marshal(log)
}

// Hash provides an md5 hash determined by Log content.
func (log Log) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(log.String()))) // #nosec
}

// Equals confirms that all fields, defined or not, in receiver Log are equal to toCompare.
func (log Log) Equals(toCompare Log) bool {
	return log.equals(toCompare, true)
}

// RelaxedEquals confirms that all defined fields in receiver Log are matched in toCompare, ignoring those not
// set.
func (log Log) RelaxedEquals(toCompare Log) bool {
	return log.equals(toCompare, false)
}

// equals determines if receiver Log is equal to toCompare Log, relaxed if not strict
func (log Log) equals(toCompare Log, strict bool) bool {
	if log.Body != toCompare.Body && (strict || log.Body != nil) {
		return false
	}
	if log.SeverityText != toCompare.SeverityText && (strict || log.SeverityText != "") {
		return false
	}

	if log.Severity != nil {
		if toCompare.Severity == nil || (*log.Severity != *toCompare.Severity) {
			return false
		}
	} else {
		if strict && toCompare.Severity != nil {
			return false
		}
	}

	return attributesAreEqual(log.Attributes, toCompare.Attributes)
}

// FlattenResourceLogs takes multiple instances of ResourceLogs and flattens them
// to only unique entries by Resource, InstrumentationScope, and Log content.
// It will preserve order by removing subsequent occurrences of repeated items
// from the returned flattened ResourceLogs item
func FlattenResourceLogs(resourceLogs ...ResourceLogs) ResourceLogs {
	flattened := ResourceLogs{}

	var resourceHashes []string
	// maps of resource hashes to objects
	resources := map[string]Resource{}
	scopeLogs := map[string][]ScopeLogs{}

	// flatten by Resource
	for _, rms := range resourceLogs {
		for _, rm := range rms.ResourceLogs {
			resourceHash := rm.Resource.Hash()
			if _, ok := resources[resourceHash]; !ok {
				resources[resourceHash] = rm.Resource
				resourceHashes = append(resourceHashes, resourceHash)
			}
			scopeLogs[resourceHash] = append(scopeLogs[resourceHash], rm.ScopeLogs...)
		}
	}

	// flatten by InstrumentationScope
	for _, resourceHash := range resourceHashes {
		resource := resources[resourceHash]
		resourceLog := ResourceLog{
			Resource: resource,
		}

		var ilHashes []string
		// maps of hashes to objects
		ils := map[string]InstrumentationScope{}
		ilLogs := map[string][]Log{}
		for _, ilm := range scopeLogs[resourceHash] {
			ilHash := ilm.Scope.Hash()
			if _, ok := ils[ilHash]; !ok {
				ils[ilHash] = ilm.Scope
				ilHashes = append(ilHashes, ilHash)
			}
			if ilm.Logs == nil {
				ilm.Logs = []Log{}
			}
			ilLogs[ilHash] = append(ilLogs[ilHash], ilm.Logs...)
		}

		// flatten by Log
		for _, ilHash := range ilHashes {
			il := ils[ilHash]

			var logHashes []string
			logs := map[string]Log{}
			allILLogs := ilLogs[ilHash]
			for _, log := range allILLogs {
				logHash := log.Hash()
				if _, ok := logs[logHash]; !ok {
					logs[logHash] = log
					logHashes = append(logHashes, logHash)
				}
			}

			var flattenedLogs []Log
			for _, logHash := range logHashes {
				flattenedLogs = append(flattenedLogs, logs[logHash])
			}

			if flattenedLogs == nil {
				flattenedLogs = []Log{}
			}

			sms := ScopeLogs{
				Scope: il,
				Logs:  flattenedLogs,
			}
			resourceLog.ScopeLogs = append(resourceLog.ScopeLogs, sms)
		}

		flattened.ResourceLogs = append(flattened.ResourceLogs, resourceLog)
	}

	return flattened
}

// ContainsAll determines if everything in expected ResourceLogs is in the receiver ResourceLogs
// item (i.e. expected ⊆ resourceLogs). Neither guarantees equivalence, nor that expected contains all of resourceLogs
// (i.e. is not an expected ≣ resourceLogs nor resourceLogs ⊆ expected check).
// Log equivalence is based on RelaxedEquals() check: fields not in expected (e.g. unit, type, value, etc.)
// are not compared to received, but all labels must match.
// For better reliability, it's advised that both ResourceLogs items have been flattened by FlattenResourceLogs.
func (resourceLogs ResourceLogs) ContainsAll(expected ResourceLogs) (bool, error) {
	var missingResources []string
	missingInstrumentationScopes := map[string]struct{}{}
	var missingLogs []string

	for _, expectedResourceLog := range expected.ResourceLogs {
		resourceMatched := false
		for _, resourceLog := range resourceLogs.ResourceLogs {
			if expectedResourceLog.Resource.Equals(resourceLog.Resource) {
				resourceMatched = true
				innerMissingInstrumentationScopes := map[string]struct{}{}
				for _, expectedSL := range expectedResourceLog.ScopeLogs {
					matchingInstrumentationScope := ""
					for _, ilm := range resourceLog.ScopeLogs {
						if expectedSL.Scope.Equals(ilm.Scope) {
							matchingInstrumentationScope = ilm.Scope.String()
							for _, expectedLog := range expectedSL.Logs {
								logFound := false
								for _, log := range ilm.Logs {
									if expectedLog.RelaxedEquals(log) {
										logFound = true
									}
								}
								if !logFound {
									missingLogs = append(missingLogs, expectedLog.String())
								}
							}
							if len(missingLogs) != 0 {
								return false, fmt.Errorf(
									"%v doesn't contain all of %v. Missing Logs: %s",
									ilm.Logs, expectedSL.Logs, missingLogs)
							}
						}
					}
					if matchingInstrumentationScope != "" {
						// no longer globally missing
						delete(missingInstrumentationScopes, matchingInstrumentationScope)
					} else {
						innerMissingInstrumentationScopes[expectedSL.Scope.String()] = struct{}{}
					}
				}
				if len(innerMissingInstrumentationScopes) != 0 {
					if expectedResourceLog.Resource.Attributes == nil {
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
						resourceLog.ScopeLogs, expectedResourceLog.ScopeLogs, missingIS,
					)
				}
			}
		}
		if !resourceMatched {
			missingResources = append(missingResources, expectedResourceLog.Resource.String())
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
			resourceLogs.ResourceLogs, expected.ResourceLogs, missingResources,
		)
	}
	return true, nil
}
