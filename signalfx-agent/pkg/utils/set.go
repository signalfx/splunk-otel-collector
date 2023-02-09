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

// UniqueStrings returns a slice with the unique set of strings from the input
func UniqueStrings(strings []string) []string {
	unique := map[string]struct{}{}
	for _, v := range strings {
		unique[v] = struct{}{}
	}

	keys := make([]string, 0)
	for k := range unique {
		keys = append(keys, k)
	}

	return keys
}

// StringSliceToMap converts a slice of strings into a map with keys from the slice
func StringSliceToMap(strings []string) map[string]bool {
	// Use bool so that the user can do `if setMap[key] { ... }``
	ret := map[string]bool{}
	for _, s := range strings {
		ret[s] = true
	}
	return ret
}

// StringSetToSlice converts a map representing a set into a slice of strings.
// If the value is `false`, the key won't be added.
func StringSetToSlice(set map[string]bool) []string {
	var out []string
	for k, ok := range set {
		if ok {
			out = append(out, k)
		}
	}
	return out
}

// MergeStringSets merges 2+ string map sets into a single output map.
func MergeStringSets(sets ...map[string]bool) map[string]bool {
	out := map[string]bool{}
	for _, ss := range sets {
		for k, v := range ss {
			out[k] = v
		}
	}
	return out
}

// StringSet creates a map set from vararg
func StringSet(strings ...string) map[string]bool {
	return StringSliceToMap(strings)
}
