// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translatesfx

import "fmt"

type component struct {
	attrs map[string]interface{}
	// baseName has the baseName for what will eventually be the key to this
	// component in the config -- e.g. "smartagent/sql" which might end up being
	// "smartagent/sql/0"
	baseName string
}

type componentCollection []component

// toComponentMap turns a componentCollection into a map such that its keys have a `/<number>`
// suffix for any components with colliding provisional keys
func (cc componentCollection) toComponentMap() map[string]map[string]interface{} {
	keyCounts := map[string]int{}
	hasMultiKeys := map[string]struct{}{}
	for _, c := range cc {
		count := keyCounts[c.baseName]
		if count > 0 {
			hasMultiKeys[c.baseName] = struct{}{}
		}
		keyCounts[c.baseName] = count + 1
	}
	keyCounts = map[string]int{}
	out := map[string]map[string]interface{}{}
	for _, c := range cc {
		_, found := hasMultiKeys[c.baseName]
		key := c.baseName
		if found {
			numSeen := keyCounts[c.baseName]
			key = fmt.Sprintf("%s/%d", key, numSeen)
			keyCounts[c.baseName] = numSeen + 1
		}
		out[key] = c.attrs
	}
	return out
}

func saMonitorToStandardReceiver(monitor map[string]interface{}) component {
	if excludes, ok := monitor[metricsToExclude]; ok {
		delete(monitor, metricsToExclude)
		monitor["datapointsToExclude"] = excludes
	}
	return component{
		baseName: "smartagent/" + monitor["type"].(string),
		attrs:    monitor,
	}
}
