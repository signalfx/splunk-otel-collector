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
	"fmt"
	"regexp"
	"strings"
)

// RegexpGroupMap matches text against the given regexp and returns a map of
// all of the named subgroups to the values found in text.  Returns nil if text
// does not match.
func RegexpGroupMap(re *regexp.Regexp, text string) map[string]string {
	if groups := re.FindStringSubmatch(text); groups != nil {
		groupMap := map[string]string{}
		for i, name := range re.SubexpNames() {
			groupMap[name] = groups[i]
		}
		return groupMap
	}
	return nil
}

// FindMatchString compares a string to an array of regular expressions and returns
// whether the string matches any of the expressions
func FindMatchString(in string, regexps []*regexp.Regexp) bool {
	for _, r := range regexps {
		if r.MatchString(in) {
			return true
		}
	}
	return false
}

// RegexpStringsToRegexp - Converts an array of strings formatted with "/.../" to an array of *regexp.Regexp
// or and returns any plain strings as a map[string]struct{}
func RegexpStringsToRegexp(regexpStrings []string) ([]*regexp.Regexp, map[string]struct{}, []error) {
	regexes := make([]*regexp.Regexp, 0, len(regexpStrings))
	strs := make(map[string]struct{}, len(regexpStrings))
	errors := []error{}
	// compile mountpoint regexes
	for _, r := range regexpStrings {
		var regexString string
		if strings.HasPrefix(r, "/") && strings.HasSuffix(r, "/") {
			regexString = strings.TrimPrefix(strings.TrimSuffix(r, "/"), "/")
			exp, err := regexp.Compile(regexString)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to compile regexp '%s'", r))
				continue
			}
			regexes = append(regexes, exp)
		} else {
			strs[r] = struct{}{}
		}

	}
	return regexes, strs, errors
}
