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

import "sort"

type otelCfg struct {
	Extensions    map[string]interface{} `yaml:",omitempty"`
	ConfigSources map[string]interface{} `yaml:"config_sources"`
	Receivers     map[string]interface{}
	Processors    map[string]interface{} `yaml:",omitempty"`
	Exporters     map[string]interface{}
	Service       map[string]interface{}
}

func saInfoToOtelConfig(cfg saCfgInfo) otelCfg {
	receivers := map[string]interface{}{}
	for _, v := range cfg.monitors {
		monitor := v.(map[interface{}]interface{})
		receiver := saMonitorToOtelReceiver(monitor)
		for k, v := range receiver {
			receivers[k.(string)] = v
		}
	}
	const resourceDetection = "resourcedetection"
	out := otelCfg{
		ConfigSources: map[string]interface{}{
			"include": nil,
		},
		Receivers: receivers,
		Processors: map[string]interface{}{
			resourceDetection: map[string]interface{}{
				"detectors": []string{"system", "env", "gce", "ecs", "ec2", "azure"},
			},
		},
		Exporters: sfxExporter(cfg.accessToken, cfg.realm),
	}
	const sfx = "signalfx"
	metricsPipeline := rpe{
		Receivers:  receiverList(receivers),
		Processors: []string{resourceDetection},
		Exporters:  []string{sfx},
	}
	if cfg.globalDims != nil {
		const metricsTransform = "metricstransform"
		out.Processors[metricsTransform] = dimsToMetricsTransformProcessor(cfg.globalDims)
		metricsPipeline.Processors = append(metricsPipeline.Processors, metricsTransform)
	}
	pipelines := map[string]interface{}{"metrics": metricsPipeline}
	out.Service = map[string]interface{}{
		"pipelines": pipelines,
	}
	if len(cfg.saExtension) > 0 {
		out.Extensions = cfg.saExtension
		out.Service["extensions"] = []string{"smartagent"}
	}
	const sfxFwder = "smartagent/signalfx-forwarder"
	if _, ok := receivers[sfxFwder]; ok {
		pipelines["traces"] = rpe{
			Receivers:  []string{sfxFwder},
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}
	return out
}

// rpe == Receivers Processors Exporters. Using this instead of a map for
// deterministic ordering.
type rpe struct {
	Receivers  []string
	Processors []string `yaml:",omitempty"`
	Exporters  []string
}

func receiverList(receivers map[string]interface{}) []string {
	var keys []string
	for k := range receivers {
		keys = append(keys, k)
	}
	return keys
}

func saMonitorToOtelReceiver(monitor map[interface{}]interface{}) map[interface{}]interface{} {
	// TODO translate discovery rule (delete for now)
	delete(monitor, "discoveryRule")
	return map[interface{}]interface{}{
		"smartagent/" + monitor["type"].(string): monitor,
	}
}

func sfxExporter(accessToken, realm string) map[string]interface{} {
	return map[string]interface{}{
		"signalfx": map[interface{}]interface{}{
			"access_token": accessToken,
			"realm":        realm,
		},
	}
}

func dimsToMetricsTransformProcessor(m map[interface{}]interface{}) map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"transforms": []map[interface{}]interface{}{{
			"include":    ".*",
			"match_type": "regexp",
			"action":     "update",
			"operations": mtOperations(m),
		}},
	}
}

func mtOperations(m map[interface{}]interface{}) (out []map[interface{}]interface{}) {
	var keys []string
	for k := range m {
		keys = append(keys, k.(string))
	}
	// sorted for easier testing
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, map[interface{}]interface{}{
			"action":    "add_label",
			"new_label": k,
			"new_value": m[k],
		})
	}
	return
}
