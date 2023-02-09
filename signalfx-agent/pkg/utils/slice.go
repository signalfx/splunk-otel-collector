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

// MakeRange creates an int slice containing all ints between `min` and `max`
func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

// InterfaceSliceToStringSlice returns a new slice that contains the elements
// of `is` as strings.  Returns nil if any of the elements of `is` are not
// strings.
func InterfaceSliceToStringSlice(is []interface{}) []string {
	var ss []string
	for _, intf := range is {
		if s, ok := intf.(string); ok {
			ss = append(ss, s)
		} else {
			return nil
		}
	}
	return ss
}

// RemoveAllElementsFromStringSlice removes all elements from toRemove that exists
// in inputStrings
func RemoveAllElementsFromStringSlice(inputStrings []string, toRemoveStrings []string) []string {
	inputStringsMap := StringSliceToMap(inputStrings)
	toRemoveStringsMap := StringSliceToMap(toRemoveStrings)

	for key := range toRemoveStringsMap {
		delete(inputStringsMap, key)
	}

	return StringSetToSlice(inputStringsMap)
}
