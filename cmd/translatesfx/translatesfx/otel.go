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
	"fmt"
	"sort"
	"strings"
)

const (
	processlist       = "smartagent/processlist"
	sfxFwder          = "smartagent/signalfx-forwarder"
	resourceDetection = "resourcedetection"
	sfx               = "signalfx"
	metricsTransform  = "metricstransform"
)

func saInfoToOtelConfig(sa saCfgInfo, vaultPaths []string) *otelCfg {
	otel := newOtelCfg()
	translateExporters(sa, otel)
	translateMonitors(sa, otel)
	translateGlobalDims(sa, otel)
	translateSAExtension(sa, otel)
	translateObservers(sa, otel)
	translateConfigSources(sa, otel, vaultPaths)
	translateFilters(sa, otel)
	return otel
}

func newOtelCfg() *otelCfg {
	return &otelCfg{
		ConfigSources: map[string]interface{}{
			"include": nil,
		},
		Receivers: map[string]map[string]interface{}{},
		Processors: map[string]map[string]interface{}{
			resourceDetection: {
				"detectors": []string{"system", "env", "gce", "ecs", "ec2", "azure"},
			},
		},
		Extensions: map[string]map[string]interface{}{},
		Service: service{
			Pipelines: map[string]*rpe{
				"metrics": {
					Processors: []string{resourceDetection},
					Exporters:  []string{sfx},
				},
				"logs": {
					Receivers:  []string{processlist},
					Processors: []string{resourceDetection},
					Exporters:  []string{sfx},
				},
			},
		},
	}
}

func translateFilters(sa saCfgInfo, otel *otelCfg) {
	if sa.metricsToExclude == nil {
		return
	}
	procName := "filter"
	otel.Processors[procName] = map[string]interface{}{
		"metrics": map[string]interface{}{
			"exclude": map[string]interface{}{
				"match_type":  "expr",
				"expressions": saFiltersToExpr(sa.metricsToExclude),
			},
		},
	}
	otel.Service.Pipelines["metrics"].appendProcessor(procName)
}

func saFiltersToExpr(excludes []interface{}) []string {
	var out []string
	for _, excludeV := range excludes {
		exclude, ok := excludeV.(map[interface{}]interface{})
		if !ok {
			return nil
		}
		var names []interface{}
		namesV, ok := exclude["metricNames"]
		if !ok {
			if nameV, metricNameOK := exclude["metricName"]; metricNameOK {
				names = []interface{}{nameV}
			} else {
				return nil
			}
		} else {
			names = namesV.([]interface{})
		}
		line := metricNamesToExpr(names)

		dimsV, ok := exclude["dimensions"]
		var dims map[interface{}]interface{}
		if ok && dimsV != nil {
			dims = dimsV.(map[interface{}]interface{})
		}
		dimExpr := dimsToExpr(dims)
		if dimExpr != "" {
			if line != "" {
				line += " and "
			}
			line += "(" + dimExpr + ")"
		}

		out = append(out, line)
	}
	return out
}

func metricNamesToExpr(names []interface{}) string {
	out := ""
	for _, nameV := range names {
		name := nameV.(string)
		rex, negated := globToRegexpStr(name)
		stmt := fmt.Sprintf("MetricName matches %q", rex)
		out += wrapStatement(stmt, out == "", negated)
	}
	return out
}

func dimsToExpr(dimSets map[interface{}]interface{}) string {
	if dimSets == nil {
		return ""
	}
	out := ""
	for dimsKey, ds := range dimSets {
		var dimSet []interface{}
		switch t := ds.(type) {
		case string:
			dimSet = []interface{}{t}
		case []interface{}:
			dimSet = t
		}

		for _, dim := range dimSet {
			rex, negated := globToRegexpStr(dim.(string))
			stmt := fmt.Sprintf("Label(%q) matches %q", dimsKey, rex)
			out += wrapStatement(stmt, out == "", negated)
		}
	}
	return out
}

func wrapStatement(stmt string, empty, negated bool) string {
	if negated {
		stmt = "not (" + stmt + ")"
		if empty {
			return stmt
		}
		return " and " + stmt
	}
	if !empty {
		return " or " + stmt
	}
	return stmt
}

func globToRegexpStr(s string) (string, bool) {
	negated := false
	if strings.HasPrefix(s, "!") {
		s = s[1:]
		negated = true
	}
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "*", ".*")
	s = strings.ReplaceAll(s, "?", ".{1}")
	return "^" + s + "$", negated
}

func translateExporters(sa saCfgInfo, cfg *otelCfg) {
	cfg.Exporters = sfxExporter(sa)
}

func translateMonitors(sa saCfgInfo, cfg *otelCfg) {
	rcReceivers := map[string]map[string]interface{}{}
	for _, monV := range sa.monitors {
		monitor := monV.(map[interface{}]interface{})
		receiver, isRC := saMonitorToOtelReceiver(monitor)
		target := cfg.Receivers
		if isRC {
			target = rcReceivers
		}
		for k, v := range receiver {
			target[k] = v
		}
	}
	mp := cfg.Service.Pipelines["metrics"]
	mp.Receivers = receiverList(cfg.Receivers)
	if len(rcReceivers) > 0 {
		const rc = "receiver_creator"
		cfg.Receivers[rc] = map[string]interface{}{
			"receivers":       rcReceivers,
			"watch_observers": []string{"k8s_observer"}, // TODO check observer type?
		}
		mp.appendReceiver(rc)
	}

	if _, ok := cfg.Receivers[sfxFwder]; ok {
		cfg.Service.Pipelines["traces"] = &rpe{
			Receivers:  []string{sfxFwder},
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}
}

func translateGlobalDims(sa saCfgInfo, otel *otelCfg) {
	if sa.globalDims != nil {
		otel.Processors[metricsTransform] = dimsToMetricsTransformProcessor(sa.globalDims)
		metricsPipeline := otel.Service.Pipelines["metrics"]
		metricsPipeline.Processors = append(metricsPipeline.Processors, metricsTransform)
	}
}

func translateSAExtension(sa saCfgInfo, otel *otelCfg) {
	if len(sa.saExtension) > 0 {
		otel.addExtensions(sa.saExtension)
		otel.Service.appendExtension("smartagent")
	}
}

func translateObservers(sa saCfgInfo, otel *otelCfg) {
	if len(sa.observers) > 0 {
		m := saObserversToOtel(sa.observers)
		if m != nil {
			otel.addExtensions(m)
			otel.Service.appendExtension("k8s_observer")
		}
	}
}

func translateConfigSources(sa saCfgInfo, otel *otelCfg, vaultPaths []string) {
	otel.ConfigSources = map[string]interface{}{
		"include": nil,
	}
	if sa.configSources == nil {
		return
	}

	translateZK(sa, otel)
	translateEtcd(sa, otel)
	translateVault(sa, otel, vaultPaths)
}

func translateZK(sa saCfgInfo, otel *otelCfg) {
	v, ok := sa.configSources["zookeeper"]
	if !ok {
		return
	}
	zk, ok := v.(map[interface{}]interface{})
	if !ok {
		return
	}
	m := map[string]interface{}{
		"endpoints": zk["endpoints"],
	}
	if tos, ok := zk["timeoutSeconds"]; ok {
		m["timeout"] = fmt.Sprintf("%ds", tos)
	}
	otel.ConfigSources["zookeeper"] = m
}

func translateEtcd(sa saCfgInfo, otel *otelCfg) {
	v, ok := sa.configSources["etcd2"]
	if !ok {
		return
	}
	etcd, o := v.(map[interface{}]interface{})
	if !o {
		return
	}
	m := map[string]interface{}{
		"endpoints": etcd["endpoints"],
	}
	auth := map[string]interface{}{}
	if username, ok := etcd["username"]; ok {
		auth["username"] = username
	}
	if password, ok := etcd["password"]; ok {
		auth["password"] = password
	}
	if len(auth) > 0 {
		m["auth"] = auth
	}
	otel.ConfigSources["etcd2"] = m
}

func translateVault(sa saCfgInfo, otel *otelCfg, vaultPaths []string) {
	if v, ok := sa.configSources["vault"]; ok {
		vault, ok := v.(map[interface{}]interface{})
		if !ok {
			return
		}
		for i, vaultPath := range vaultPaths {
			otel.ConfigSources[fmt.Sprintf("vault/%d", i)] = map[string]interface{}{
				"endpoint": vault["vaultAddr"],
				"path":     vaultPath,
				"auth": map[string]interface{}{
					"token": vault["vaultToken"],
				},
			}
		}
	}
}

func dimsToMetricsTransformProcessor(m map[interface{}]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"transforms": []map[interface{}]interface{}{{
			"include":    ".*",
			"match_type": "regexp",
			"action":     "update",
			"operations": mtOperations(m),
		}},
	}
}

func receiverList(receivers map[string]map[string]interface{}) []string {
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

func sfxExporter(sa saCfgInfo) map[string]map[string]interface{} {
	cfg := map[string]interface{}{
		"access_token": sa.accessToken,
	}
	if sa.realm != "" {
		cfg["realm"] = sa.realm
	}
	if sa.ingestURL != "" {
		cfg["ingest_url"] = sa.ingestURL
	}
	if sa.APIURL != "" {
		cfg["api_url"] = sa.APIURL
	}
	return map[string]map[string]interface{}{
		"signalfx": cfg,
	}
}

func saMonitorToOtelReceiver(monitor map[interface{}]interface{}) (map[string]map[string]interface{}, bool) {
	strm := interfaceMapToStringMap(monitor)
	if _, ok := monitor["discoveryRule"]; ok {
		return saMonitorToRCReceiver(strm), true
	}
	return saMonitorToStandardReceiver(strm), false
}

func interfaceMapToStringMap(in map[interface{}]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range in {
		out[k.(string)] = v
	}
	return out
}

func stringMapToInterfaceMap(in map[string]interface{}) map[interface{}]interface{} {
	out := map[interface{}]interface{}{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func saMonitorToRCReceiver(monitor map[string]interface{}) map[string]map[string]interface{} {
	key := "smartagent/" + monitor["type"].(string)
	rcr := discoveryRuleToRCRule(monitor["discoveryRule"].(string))
	delete(monitor, "discoveryRule")
	return map[string]map[string]interface{}{
		key: {
			"rule":   rcr,
			"config": monitor,
		},
	}
}

func saMonitorToStandardReceiver(monitor map[string]interface{}) map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"smartagent/" + monitor["type"].(string): monitor,
	}
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

func discoveryRuleToRCRule(dr string) string {
	out := strings.ReplaceAll(dr, "=~", "matches")
	out = strings.ReplaceAll(out, "container_image", "pod.name")
	if strings.Contains(out, "port") {
		out = `type == "port" && ` + out
	}
	return out
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

type otelCfg struct {
	ConfigSources map[string]interface{}            `yaml:"config_sources"`
	Extensions    map[string]map[string]interface{} `yaml:",omitempty"`
	Receivers     map[string]map[string]interface{}
	Processors    map[string]map[string]interface{} `yaml:",omitempty"`
	Exporters     map[string]map[string]interface{}
	Service       service
}

func (c *otelCfg) addExtensions(m map[string]interface{}) {
	for k, v := range m {
		c.Extensions[k] = v.(map[string]interface{})
	}
}

type service struct {
	Pipelines  map[string]*rpe
	Extensions []string
}

func (s *service) appendExtension(ext string) {
	s.Extensions = append(s.Extensions, ext)
	sort.Strings(s.Extensions)
}

// rpe == Receivers Processors Exporters
type rpe struct {
	Receivers  []string
	Processors []string `yaml:",omitempty"`
	Exporters  []string
}

func (r *rpe) appendReceiver(name string) {
	r.Receivers = append(r.Receivers, name)
	sort.Strings(r.Receivers)
}

func (r *rpe) appendProcessor(name string) {
	r.Processors = append(r.Processors, name)
	sort.Strings(r.Processors)
}
