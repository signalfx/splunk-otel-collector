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

import (
	"sort"
	"strings"
)

type otelCfg struct {
	Extensions    map[string]interface{} `yaml:",omitempty"`
	ConfigSources map[string]interface{} `yaml:"config_sources"`
	Receivers     map[string]interface{}
	Processors    map[string]interface{} `yaml:",omitempty"`
	Exporters     map[string]interface{}
	Service       map[string]interface{}
}

const processlist = "smartagent/processlist"

func saInfoToOtelConfig(cfg saCfgInfo) otelCfg {
	receivers := map[string]interface{}{}
	rcReceivers := map[string]interface{}{}
	for _, monV := range cfg.monitors {
		monitor := monV.(map[interface{}]interface{})
		receiver, isRC := saMonitorToOtelReceiver(monitor)
		target := receivers
		if isRC {
			target = rcReceivers
		}
		for k, v := range receiver {
			target[k.(string)] = v
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
		Exporters:  sfxExporter(cfg.accessToken, cfg.realm),
		Extensions: map[string]interface{}{},
	}
	if len(rcReceivers) > 0 {
		out.Receivers["receiver_creator"] = map[string]interface{}{
			"receivers":       rcReceivers,
			"watch_observers": []string{"k8s_observer"}, // TODO check observer type?
		}
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
		appendMap(out.Extensions, cfg.saExtension)
		appendExtensions(out.Service, "smartagent")
	}
	if len(cfg.observers) > 0 {
		m := saObserversToOtel(cfg.observers)
		if m != nil {
			appendMap(out.Extensions, m)
			appendExtensions(out.Service, "k8s_observer")
		}
	}
	const sfxFwder = "smartagent/signalfx-forwarder"
	if _, ok := receivers[sfxFwder]; ok {
		pipelines["traces"] = rpe{
			Receivers:  []string{sfxFwder},
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}
	if _, ok := receivers[processlist]; ok {
		pipelines["logs"] = rpe{
			Receivers:  []string{processlist},
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}
	return out
}

func appendExtensions(m map[string]interface{}, v string) {
	const k = "extensions"
	_, ok := m[k]
	if !ok {
		m[k] = []string{v}
		return
	}
	m[k] = append(m[k].([]string), v)
	sort.Strings(m[k].([]string))
}

func saObserversToOtel(observers []interface{}) map[string]interface{} {
	for _, v := range observers {
		obs, ok := v.(map[interface{}]interface{})
		if !ok {
			return nil
		}
		typeV, ok := obs["type"]
		if !ok {
			return nil
		}
		observerType, ok := typeV.(string)
		if !ok {
			return nil
		}
		if observerType == "k8s-api" {
			return map[string]interface{}{
				"k8s_observer": map[string]interface{}{
					"auth_type": "serviceAccount",
					"node":      "${K8S_NODE_NAME}",
				},
			}
		}
	}
	return nil
}

func appendMap(m, n map[string]interface{}) {
	for k, v := range n {
		m[k] = v
	}
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
		if k == processlist {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func saMonitorToOtelReceiver(monitor map[interface{}]interface{}) (map[interface{}]interface{}, bool) {
	if _, ok := monitor["discoveryRule"]; ok {
		return saMonitorToRCReceiver(monitor), true
	}
	return saMonitorToStandardReceiver(monitor), false
}

func saMonitorToRCReceiver(monitor map[interface{}]interface{}) map[interface{}]interface{} {
	key := "smartagent/" + monitor["type"].(string)
	rcr := discoveryRuleToRCRule(monitor["discoveryRule"].(string))
	delete(monitor, "discoveryRule")
	out := map[interface{}]interface{}{
		key: map[string]interface{}{
			"rule":   rcr,
			"config": monitor,
		},
	}
	return out
}

func discoveryRuleToRCRule(dr string) string {
	out := strings.ReplaceAll(dr, "=~", "matches")
	out = strings.ReplaceAll(out, "container_image", "pod.name")
	if strings.Contains(out, "port") {
		out = `type == "port" && ` + out
	}
	return out
}

func saMonitorToStandardReceiver(monitor map[interface{}]interface{}) map[interface{}]interface{} {
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
