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

type saCfgInfo struct {
	accessToken      string
	realm            string
	ingestURL        string
	APIURL           string
	globalDims       map[interface{}]interface{}
	saExtension      map[string]interface{}
	configSources    map[interface{}]interface{}
	writer           map[interface{}]interface{}
	monitors         []interface{}
	observers        []interface{}
	metricsToExclude []interface{}
	metricsToInclude []interface{}
}

func saExpandedToCfgInfo(saExpanded map[interface{}]interface{}) saCfgInfo {
	return saCfgInfo{
		accessToken:      toString(saExpanded, "signalFxAccessToken"),
		realm:            toString(saExpanded, "signalFxRealm"),
		ingestURL:        toString(saExpanded, "ingestUrl"),
		APIURL:           toString(saExpanded, "apiUrl"),
		monitors:         saExpanded["monitors"].([]interface{}),
		globalDims:       globalDims(saExpanded),
		saExtension:      saExtension(saExpanded),
		observers:        observers(saExpanded),
		configSources:    configSources(saExpanded),
		metricsToExclude: saToMetricsToExclude(saExpanded),
		metricsToInclude: saToMetricsToInclude(saExpanded),
		writer:           writer(saExpanded),
	}
}

func toString(saExpanded map[interface{}]interface{}, name string) string {
	v, ok := saExpanded[name]
	if !ok {
		return ""
	}
	return v.(string)
}

func observers(saExpanded map[interface{}]interface{}) []interface{} {
	v, ok := saExpanded["observers"]
	if !ok {
		return nil
	}
	obs, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return obs
}

func globalDims(saExpanded map[interface{}]interface{}) map[interface{}]interface{} {
	var out map[interface{}]interface{}
	if gd, ok := saExpanded["globalDimensions"]; ok {
		out = gd.(map[interface{}]interface{})
	}
	return out
}

func saExtension(saExpanded map[interface{}]interface{}) map[string]interface{} {
	keys := []string{
		"bundleDir",
		"procPath",
		"etcPath",
		"varPath",
		"runPath",
		"sysPath",
		"collectd",
	}
	extensionAttrs := map[string]interface{}{}
	for _, key := range keys {
		if v, ok := saExpanded[key]; ok {
			extensionAttrs[key] = v
		}
	}
	if len(extensionAttrs) == 0 {
		return nil
	}
	return map[string]interface{}{
		"smartagent": extensionAttrs,
	}
}

func configSources(saExpanded map[interface{}]interface{}) map[interface{}]interface{} {
	v, ok := saExpanded["configSources"]
	if !ok {
		return nil
	}
	cs, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil
	}
	return cs
}

func saToMetricsToExclude(saExpanded map[interface{}]interface{}) []interface{} {
	return toSlice(saExpanded, "metricsToExclude")
}

func saToMetricsToInclude(saExpanded map[interface{}]interface{}) []interface{} {
	return toSlice(saExpanded, "metricsToInclude")
}

func toSlice(saExpanded map[interface{}]interface{}, key string) []interface{} {
	v, ok := saExpanded[key]
	if !ok {
		return nil
	}
	l, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return l
}

func writer(saExpanded map[interface{}]interface{}) map[interface{}]interface{} {
	v, ok := saExpanded["writer"]
	if !ok {
		return nil
	}
	wr, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil
	}
	return wr
}
