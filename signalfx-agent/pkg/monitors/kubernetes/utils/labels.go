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
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var propNameSanitizer = strings.NewReplacer(
	".", "_",
	"/", "_")

// PropsAndTagsFromLabels converts k8s label set into SignalFx
// properties and tags formatted sets.
func PropsAndTagsFromLabels(labels map[string]string) (map[string]string, map[string]bool) {
	props := make(map[string]string)
	tags := make(map[string]bool)

	for label, value := range labels {
		key := propNameSanitizer.Replace(label)
		// K8s labels without values are treated as tags
		if value == "" {
			tags[key] = true
		} else {
			props[key] = value
		}
	}

	return props, tags
}

func SelectorMatchesPod(selector map[string]string, pod *v1.Pod) bool {
	return labels.Set(selector).AsSelectorPreValidated().Matches(labels.Set(pod.Labels))
}
