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
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	processlist       = "smartagent/processlist"
	kubernetesEvents  = "smartagent/kubernetes-events"
	sfxFwder          = "smartagent/signalfx-forwarder"
	resourceDetection = "resourcedetection"
	sfx               = "signalfx"
	sapm              = "sapm"
	metricsTransform  = "metricstransform"
	filterProc        = "filter"
	k8sObserver       = "k8s_observer"
	hostObserver      = "host_observer"
	metricsToExclude  = "metricsToExclude"
	discoveryRule     = "discoveryRule"
)

// monitors that should be converted to both metrics and traces
var metricsAndTracesReceiverMonitorTypes = map[string]bool{sfxFwder: true}

// monitors that should be converted to logs receivers only
var exclusivelyLogsReceiverMonitorTypes = map[string]bool{processlist: true, kubernetesEvents: true}

func saInfoToOtelConfig(sa saCfgInfo, vaultPaths []string) (otel *otelCfg, warnings []error) {
	otel = newOtelCfg()
	translateExporters(sa, otel)
	w := translateMonitors(sa, otel)
	warnings = append(warnings, w...)
	translateGlobalDims(sa, otel)
	translateSAExtension(sa, otel)
	translateObservers(sa, otel)
	translateConfigSources(sa, otel, vaultPaths)
	translateFilters(sa, otel)
	return otel, warnings
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
			Pipelines: map[string]*rpe{},
		},
	}
}

func translateFilters(sa saCfgInfo, otel *otelCfg) {
	if sa.metricsToExclude == nil && sa.metricsToInclude == nil {
		return
	}

	metricsFilter := map[string]interface{}{}
	otel.Processors[filterProc] = map[string]interface{}{
		"metrics": metricsFilter,
	}

	excludeExpressions := saExcludesToExpr(sa.metricsToExclude, sa.metricsToInclude, false)
	if excludeExpressions != nil {
		metricsFilter["exclude"] = map[string]interface{}{
			"match_type":  "expr",
			"expressions": excludeExpressions,
		}
	}

	excludeExpressionsNegated := saExcludesToExpr(sa.metricsToExclude, sa.metricsToInclude, true)
	if excludeExpressionsNegated != nil {
		metricsFilter["include"] = map[string]interface{}{
			"match_type":  "expr",
			"expressions": excludeExpressionsNegated,
		}
	}

	otel.Service.Pipelines["metrics"].appendProcessor(filterProc)
}

func saExcludesToExpr(excludes []interface{}, overrides []interface{}, expectedNegation bool) []string {
	overridesExpr := saIncludesToExpr(overrides)
	var out []string
	for _, v := range excludes {
		line := filterToExpr(v.(map[interface{}]interface{}), false, expectedNegation)
		if line == "" {
			continue
		}
		if overridesExpr != "" {
			line += " and (" + overridesExpr + ")"
		}
		out = append(out, line)
	}
	return out
}

func saIncludesToExpr(includes []interface{}) string {
	out := ""
	for _, includeV := range includes {
		line := filterToExpr(includeV.(map[interface{}]interface{}), true, false)
		if line == "" {
			continue
		}
		if out != "" {
			out += " and "
		}
		out += line
	}
	return out
}

func filterToExpr(filter map[interface{}]interface{}, flipNegation, expectedNegation bool) string {
	effectiveNegated := false
	if negatedV, ok := filter["negated"]; ok {
		effectiveNegated = negatedV.(bool)
	}

	if effectiveNegated != expectedNegation {
		return ""
	}

	var names []interface{}
	if namesV, ok := filter["metricNames"]; ok {
		names = namesV.([]interface{})
	} else if nameV, metricNameOK := filter["metricName"]; metricNameOK {
		names = []interface{}{nameV}
	}
	line := metricNamesToExpr(names, flipNegation)
	dimsV, ok := filter["dimensions"]
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
	return line
}

func metricNamesToExpr(names []interface{}, flipNegation bool) string {
	out := ""
	for _, nameV := range names {
		name := nameV.(string)
		rex, negated := saToRegexpStr(name)
		stmt := fmt.Sprintf("MetricName matches %q", rex)
		if flipNegation {
			negated = !negated
		}
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
			rex, negated := saToRegexpStr(dim.(string))
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

func saToRegexpStr(s string) (string, bool) {
	negated := false
	if strings.HasPrefix(s, "!") {
		s = s[1:]
		negated = true
	}
	if isRegexFilter(s) {
		return s[1 : len(s)-1], negated
	}
	return globToRegex(s), negated
}

func isRegexFilter(s string) bool {
	return len(s) > 1 && s[0] == '/' && s[len(s)-1] == '/'
}

func globToRegex(s string) string {
	s = strings.ReplaceAll(s, ".", `\.`)
	s = strings.ReplaceAll(s, "*", ".*")
	s = strings.ReplaceAll(s, "?", ".{1}")
	return "^" + s + "$"
}

func translateExporters(sa saCfgInfo, cfg *otelCfg) {
	cfg.Exporters = sfxExporter(sa)
}

func translateMonitors(sa saCfgInfo, cfg *otelCfg) (warnings []error) {
	rcReceivers := map[string]map[string]interface{}{}
	for _, monV := range sa.monitors {
		monitor := monV.(map[interface{}]interface{})
		receiver, w, isRC := saMonitorToOtelReceiver(monitor, sa.observers)
		warnings = append(warnings, w...)
		target := cfg.Receivers
		if isRC {
			target = rcReceivers
		}
		for k, v := range receiver {
			target[k] = v
		}
	}

	metricsReceivers, tracesReceivers, logsReceivers := receiverLists(cfg.Receivers)

	if len(rcReceivers) > 0 {
		switch {
		case sa.observers == nil:
			warnings = append(warnings, errors.New("found Smart Agent discovery rule but no observers"))
		case len(sa.observers) > 1:
			warnings = append(warnings, errors.New("found Smart Agent discovery rule but multiple observers"))
		default:
			obs := saObserverTypeToOtel(sa.observers[0].(map[interface{}]interface{})["type"].(string))
			const rc = "receiver_creator"
			cfg.Receivers[rc] = map[string]interface{}{
				"receivers":       rcReceivers,
				"watch_observers": []string{obs},
			}
			metricsReceivers = append(metricsReceivers, rc)
			sort.Strings(metricsReceivers)
		}
	}

	if metricsReceivers != nil {
		cfg.Service.Pipelines["metrics"] = &rpe{
			Receivers:  metricsReceivers,
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}

	if tracesReceivers != nil {
		exporters := []string{sapm}
		if sendTraceCorrelation(sa) {
			exporters = append(exporters, sfx)
		}
		cfg.Service.Pipelines["traces"] = &rpe{
			Receivers:  tracesReceivers,
			Processors: []string{resourceDetection},
			Exporters:  exporters,
		}
		cfg.Exporters[sapm] = sapmExporter(sa)
	}

	if logsReceivers != nil {
		cfg.Service.Pipelines["logs"] = &rpe{
			Receivers:  logsReceivers,
			Processors: []string{resourceDetection},
			Exporters:  []string{sfx},
		}
	}

	return warnings
}

func sendTraceCorrelation(sa saCfgInfo) bool {
	if v, ok := sa.writer["sendTraceHostCorrelationMetrics"]; ok {
		if correlation, ok := v.(bool); ok {
			return correlation
		}
	}
	return true
}

func sapmExporter(sa saCfgInfo) map[string]interface{} {
	return map[string]interface{}{
		"access_token": sa.accessToken,
		"endpoint":     sapmEndpoint(sa),
	}
}

func sapmEndpoint(sa saCfgInfo) string {
	if sa.ingestURL != "" {
		return fmt.Sprintf("%s/v2/trace", sa.ingestURL)
	}
	if sa.realm != "" {
		return fmt.Sprintf("https://ingest.%s.signalfx.com/v2/trace", sa.realm)
	}
	return ""
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
			for k := range m {
				otel.Service.appendExtension(k)
			}
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

func receiverLists(receivers map[string]map[string]interface{}) (metrics, traces, logs []string) {
	for k := range receivers {
		if _, ok := exclusivelyLogsReceiverMonitorTypes[k]; ok {
			logs = append(logs, k)
			continue
		}

		if _, ok := metricsAndTracesReceiverMonitorTypes[k]; ok {
			traces = append(traces, k)
		}

		metrics = append(metrics, k)
	}
	sort.Strings(metrics)
	sort.Strings(traces)
	sort.Strings(logs)
	return
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

func saMonitorToOtelReceiver(monitor map[interface{}]interface{}, observers []interface{}) (
	out map[string]map[string]interface{},
	warnings []error,
	isReceiverCreator bool,
) {
	strm := interfaceMapToStringMap(monitor)
	if _, ok := monitor[discoveryRule]; ok {
		receiver, w := saMonitorToRCReceiver(strm, observers)
		return receiver, w, true
	}
	return saMonitorToStandardReceiver(strm), nil, false
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

func saMonitorToRCReceiver(monitor map[string]interface{}, observers []interface{}) (out map[string]map[string]interface{}, warnings []error) {
	key := "smartagent/" + monitor["type"].(string)
	dr := monitor[discoveryRule].(string)
	rcr, err := discoveryRuleToRCRule(dr, observers)
	if err != nil {
		// fall back to original rule if unable to translate
		rcr = dr
		warnings = append(warnings, err)
	}
	delete(monitor, discoveryRule)
	return map[string]map[string]interface{}{
		key: {
			"rule":   rcr,
			"config": monitor,
		},
	}, warnings
}

func saMonitorToStandardReceiver(monitor map[string]interface{}) map[string]map[string]interface{} {
	if excludes, ok := monitor[metricsToExclude]; ok {
		delete(monitor, metricsToExclude)
		monitor["datapointsToExclude"] = excludes
	}
	return map[string]map[string]interface{}{
		"smartagent/" + monitor["type"].(string): monitor,
	}
}

func saObserversToOtel(observers []interface{}) map[string]interface{} {
	out := map[string]interface{}{}
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
		otelObserverType := saObserverTypeToOtel(observerType)
		switch otelObserverType {
		case k8sObserver:
			out[k8sObserver] = map[string]interface{}{
				"auth_type": "serviceAccount",
				"node":      "${K8S_NODE_NAME}",
			}
		case hostObserver:
			out[hostObserver] = map[string]interface{}{}
		}
	}
	return out
}

func saObserverTypeToOtel(saType string) string {
	switch saType {
	case "k8s-api":
		return k8sObserver
	case "host":
		return hostObserver
	}
	return ""
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
