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

package telemetry

import "gopkg.in/yaml.v2"

// sanitizeAttributes helps ensure that unmarshaled yaml mappings and
// pcommon.Map items have the same map[string]any representation
// suitable for reflect.DeepEquals.
func sanitizeAttributes(attributes map[string]any) map[string]any {
	sanitized := map[string]any{}
	b, err := yaml.Marshal(attributes)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(b, &sanitized); err != nil {
		panic(err)
	}
	return sanitized
}
