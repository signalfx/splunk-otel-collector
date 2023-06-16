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
	globalDims       map[any]any
	saExtension      map[string]any
	configSources    map[any]any
	writer           map[any]any
	monitors         []any
	observers        []any
	metricsToExclude []any
	metricsToInclude []any
}

func saExpandedToCfgInfo(saExpanded map[any]any) saCfgInfo {
	return saCfgInfo{
		accessToken:      toString(saExpanded, "signalFxAccessToken"),
		realm:            toString(saExpanded, "signalFxRealm"),
		ingestURL:        toString(saExpanded, "ingestUrl"),
		APIURL:           toString(saExpanded, "apiUrl"),
		monitors:         saExpanded["monitors"].([]any),
		globalDims:       globalDims(saExpanded),
		saExtension:      saExtension(saExpanded),
		observers:        observers(saExpanded),
		configSources:    configSources(saExpanded),
		metricsToExclude: saToMetricsToExclude(saExpanded),
		metricsToInclude: saToMetricsToInclude(saExpanded),
		writer:           writer(saExpanded),
	}
}

func toString(saExpanded map[any]any, name string) string {
	v, ok := saExpanded[name]
	if !ok {
		return ""
	}
	if v == nil {
		// this can happen if an included file is empty
		return ""
	}
	return v.(string)
}

func observers(saExpanded map[any]any) []any {
	v, ok := saExpanded["observers"]
	if !ok {
		return nil
	}
	obs, ok := v.([]any)
	if !ok {
		return nil
	}
	return obs
}

func globalDims(saExpanded map[any]any) map[any]any {
	var out map[any]any
	if gd, ok := saExpanded["globalDimensions"]; ok {
		out = gd.(map[any]any)
	}
	return out
}

func saExtension(saExpanded map[any]any) map[string]any {
	keys := []string{
		"bundleDir",
		"procPath",
		"etcPath",
		"varPath",
		"runPath",
		"sysPath",
		"collectd",
	}
	extensionAttrs := map[string]any{}
	for _, key := range keys {
		if v, ok := saExpanded[key]; ok {
			extensionAttrs[key] = v
		}
	}
	if len(extensionAttrs) == 0 {
		return nil
	}
	return map[string]any{
		"smartagent": extensionAttrs,
	}
}

func configSources(saExpanded map[any]any) map[any]any {
	v, ok := saExpanded["configSources"]
	if !ok {
		return nil
	}
	cs, ok := v.(map[any]any)
	if !ok {
		return nil
	}
	return cs
}

func saToMetricsToExclude(saExpanded map[any]any) []any {
	return toSlice(saExpanded, "metricsToExclude")
}

func saToMetricsToInclude(saExpanded map[any]any) []any {
	return toSlice(saExpanded, "metricsToInclude")
}

func toSlice(saExpanded map[any]any, key string) []any {
	v, ok := saExpanded[key]
	if !ok {
		return nil
	}
	l, ok := v.([]any)
	if !ok {
		return nil
	}
	return l
}

func writer(saExpanded map[any]any) map[any]any {
	v, ok := saExpanded["writer"]
	if !ok {
		return nil
	}
	wr, ok := v.(map[any]any)
	if !ok {
		return nil
	}
	return wr
}
