// Copyright Splunk, Inc.
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

package gnmireceiver

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/otel/semconv/v1.22.0"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver/internal/metadata"
)

const (
	infoMetricSuffix  = "_info"
	stateMetricSuffix = "_state"
	infoValueAttr     = "value"
	stateValueAttr    = "state"
	indexAttr         = "index"
)

// metricParser converts gNMI SubscribeResponse messages into OTel metrics.
type metricParser struct {
	schema        *yangSchema
	logger        *zap.Logger
	unresolved    sync.Map
	endpoint      string
	subscriptions []SubscriptionConfig
}

func newMetricParser(endpoint string, subscriptions []SubscriptionConfig, schema *yangSchema, logger *zap.Logger) *metricParser {
	for i := range subscriptions {
		sub := &subscriptions[i]
		if sub.Default != nil {
			cacheNormalizedEnumValues(sub.Default)
		}
		for leaf, cfg := range sub.Overrides {
			cacheNormalizedEnumValues(&cfg)
			sub.Overrides[leaf] = cfg
		}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &metricParser{endpoint: endpoint, subscriptions: subscriptions, schema: schema, logger: logger}
}

func cacheNormalizedEnumValues(cfg *MetricConfig) {
	if len(cfg.EnumValues) == 0 {
		return
	}
	cfg.normalizedEnumValues = make([]string, len(cfg.EnumValues))
	for i, v := range cfg.EnumValues {
		cfg.normalizedEnumValues[i] = normalizeEnumValue(v)
	}
}

// parseBatch accumulates the metrics produced while converting a single SubscribeResponse.
type parseBatch struct {
	sm     pmetric.ScopeMetrics
	byName map[string]pmetric.Metric
}

func newParseBatch(sm pmetric.ScopeMetrics) *parseBatch {
	return &parseBatch{sm: sm, byName: map[string]pmetric.Metric{}}
}

func (b *parseBatch) numberDataPoint(name, unit string, wantType pmetric.MetricType) (pmetric.NumberDataPoint, error) {
	if m, ok := b.byName[name]; ok {
		if m.Type() != wantType {
			return pmetric.NumberDataPoint{}, fmt.Errorf(
				"metric %q is already emitted as %s in this batch, cannot also emit it as %s",
				name, m.Type(), wantType)
		}
		if m.Unit() != unit {
			return pmetric.NumberDataPoint{}, fmt.Errorf(
				"metric %q is already emitted with unit %q in this batch, cannot also emit it with unit %q",
				name, m.Unit(), unit)
		}
		if wantType == pmetric.MetricTypeSum {
			return m.Sum().DataPoints().AppendEmpty(), nil
		}
		return m.Gauge().DataPoints().AppendEmpty(), nil
	}

	m := b.sm.Metrics().AppendEmpty()
	m.SetName(name)
	m.SetUnit(unit)
	b.byName[name] = m

	if wantType == pmetric.MetricTypeSum {
		sum := m.SetEmptySum()
		sum.SetIsMonotonic(true)
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		return sum.DataPoints().AppendEmpty(), nil
	}
	return m.SetEmptyGauge().DataPoints().AppendEmpty(), nil
}

func (p *metricParser) parse(resp *gnmipb.SubscribeResponse) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()

	notification := resp.GetUpdate()
	if notification == nil || len(notification.GetUpdate()) == 0 {
		return metrics, nil
	}

	ts := pcommon.NewTimestampFromTime(time.Now())
	if notification.GetTimestamp() != 0 {
		ts = pcommon.Timestamp(notification.GetTimestamp()) //nolint:gosec // G115: gNMI timestamps are non-negative unix nanos
	}

	b := newParseBatch(p.newScopeMetrics(metrics))

	var errs []string
	for _, update := range notification.GetUpdate() {
		elems, keys := joinPath(notification.GetPrefix(), update.GetPath())
		if len(elems) == 0 {
			continue
		}
		origin := notification.GetPrefix().GetOrigin()
		if origin == "" {
			origin = update.GetPath().GetOrigin()
		}

		if err := p.appendUpdate(b, origin, elems, keys, update.GetVal(), ts); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(b.byName) == 0 {
		metrics = pmetric.NewMetrics()
	}
	if len(errs) > 0 {
		return metrics, fmt.Errorf("failed to convert %d update(s): %s", len(errs), strings.Join(errs, "; "))
	}
	return metrics, nil
}

func (p *metricParser) newScopeMetrics(metrics pmetric.Metrics) pmetric.ScopeMetrics {
	rm := metrics.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr(string(conventions.ServerAddressKey), p.endpoint)
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName(metadata.ScopeName)
	return sm
}

func (p *metricParser) appendUpdate(
	b *parseBatch,
	origin string,
	elems []string,
	keys map[string]string,
	val *gnmipb.TypedValue,
	ts pcommon.Timestamp,
) error {
	if val == nil {
		return nil
	}

	switch v := val.GetValue().(type) {
	case *gnmipb.TypedValue_UintVal:
		return p.emitUint(b, origin, elems, keys, v.UintVal, ts)
	case *gnmipb.TypedValue_IntVal:
		return p.emitInt(b, origin, elems, keys, v.IntVal, ts)
	case *gnmipb.TypedValue_FloatVal:
		return p.emitDouble(b, origin, elems, keys, float64(v.FloatVal), ts) //nolint:staticcheck // SA1019: float_val is deprecated but still sent by some targets
	case *gnmipb.TypedValue_DoubleVal:
		return p.emitDouble(b, origin, elems, keys, v.DoubleVal, ts)
	case *gnmipb.TypedValue_BoolVal:
		var n int64
		if v.BoolVal {
			n = 1
		}
		return p.emitInt(b, origin, elems, keys, n, ts)
	case *gnmipb.TypedValue_StringVal:
		return p.emitInfo(b, origin, elems, keys, v.StringVal, ts)
	case *gnmipb.TypedValue_JsonVal:
		return p.emitJSON(b, origin, elems, keys, v.JsonVal, ts)
	case *gnmipb.TypedValue_JsonIetfVal:
		return p.emitJSON(b, origin, elems, keys, v.JsonIetfVal, ts)
	case *gnmipb.TypedValue_LeaflistVal:
		var errs []string
		for i, element := range v.LeaflistVal.GetElement() {
			if err := p.appendUpdate(b, origin, elems, withIndex(keys, i), element, ts); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return errors.New(strings.Join(errs, "; "))
		}
		return nil
	default:
		return nil
	}
}

func withIndex(attrs map[string]string, i int) map[string]string {
	indexed := make(map[string]string, len(attrs)+1)
	for k, v := range attrs {
		indexed[k] = v
	}
	indexed[indexAttr] = strconv.Itoa(i)
	return indexed
}

func (p *metricParser) emitUint(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, value uint64, ts pcommon.Timestamp,
) error {
	if value > math.MaxInt64 {
		if err := p.emitDouble(b, origin, elems, keys, float64(value), ts); err != nil {
			return err
		}
		return fmt.Errorf("value %d for %q exceeds int64 range; emitted as a double and lost precision",
			value, buildMetricName(origin, elems))
	}
	return p.emitInt(b, origin, elems, keys, int64(value), ts)
}

func (p *metricParser) emitInt(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, value int64, ts pcommon.Timestamp,
) error {
	rm, ok := p.resolve(origin, elems)
	if !ok {
		return nil
	}
	return p.writeInt(b, origin, elems, keys, rm.MetricConfig, value, ts)
}

func (p *metricParser) emitDouble(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, value float64, ts pcommon.Timestamp,
) error {
	rm, ok := p.resolve(origin, elems)
	if !ok {
		return nil
	}
	return p.writeDouble(b, origin, elems, keys, rm.MetricConfig, value, ts)
}

func (p *metricParser) emitInfo(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, value string, ts pcommon.Timestamp,
) error {
	rm, ok := p.resolve(origin, elems)
	if !ok {
		return nil
	}
	cfg := rm.MetricConfig

	if len(cfg.EnumValues) > 0 {
		return p.writeEnumState(b, origin, elems, keys, cfg, value, ts)
	}

	switch rm.kind {
	case valueKindString:
		// The schema says this leaf is a string; never coerce it to a number.
	case valueKindInt:
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			return p.writeInt(b, origin, elems, keys, cfg, n, ts)
		}
	case valueKindFloat:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return p.writeDouble(b, origin, elems, keys, cfg, f, ts)
		}
	case valueKindUnknown:
		if cfg.Type == metricTypeSum || cfg.Type == metricTypeGauge {
			if n, err := strconv.ParseInt(value, 10, 64); err == nil {
				return p.writeInt(b, origin, elems, keys, cfg, n, ts)
			}
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				return p.writeDouble(b, origin, elems, keys, cfg, f, ts)
			}
		}
	}

	return p.writeInfo(b, origin, elems, keys, cfg, value, ts)
}

func (p *metricParser) writeInt(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, cfg MetricConfig, value int64, ts pcommon.Timestamp,
) error {
	dp, err := b.numberDataPoint(buildMetricName(origin, elems), cfg.Unit, metricDataType(cfg.Type))
	if err != nil {
		return err
	}
	dp.SetIntValue(value)
	dp.SetTimestamp(ts)
	putAttrs(dp.Attributes(), keys)
	return nil
}

func (p *metricParser) writeDouble(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, cfg MetricConfig, value float64, ts pcommon.Timestamp,
) error {
	dp, err := b.numberDataPoint(buildMetricName(origin, elems), cfg.Unit, metricDataType(cfg.Type))
	if err != nil {
		return err
	}
	dp.SetDoubleValue(value)
	dp.SetTimestamp(ts)
	putAttrs(dp.Attributes(), keys)
	return nil
}

func (p *metricParser) writeInfo(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, _ MetricConfig, value string, ts pcommon.Timestamp,
) error {
	name := buildMetricName(origin, elems) + infoMetricSuffix
	dp, err := b.numberDataPoint(name, "", pmetric.MetricTypeGauge)
	if err != nil {
		return err
	}
	dp.SetIntValue(1)
	dp.SetTimestamp(ts)
	dp.Attributes().PutStr(infoValueAttr, value)
	putAttrs(dp.Attributes(), keys)
	return nil
}

func (p *metricParser) writeEnumState(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, cfg MetricConfig, value string, ts pcommon.Timestamp,
) error {
	metricName := buildMetricName(origin, elems) + stateMetricSuffix
	unit := cfg.Unit
	if unit == "" {
		unit = "1"
	}

	normalized := normalizeEnumValue(value)
	matched := false
	for _, normalizedEnumValue := range cfg.normalizedEnumValues {
		var n int64
		if normalizedEnumValue == normalized {
			n = 1
			matched = true
		}

		dp, err := b.numberDataPoint(metricName, unit, pmetric.MetricTypeGauge)
		if err != nil {
			return err
		}
		dp.SetIntValue(n)
		dp.SetTimestamp(ts)
		dp.Attributes().PutStr(stateValueAttr, normalizedEnumValue)
		putAttrs(dp.Attributes(), keys)
	}

	if !matched {
		return fmt.Errorf("value %q for %q is not in enum_values %v; emitted all states as 0",
			value, metricName, cfg.EnumValues)
	}
	return nil
}

func normalizeEnumValue(value string) string {
	if idx := strings.LastIndexByte(value, ':'); idx >= 0 {
		value = value[idx+1:]
	}
	return strings.TrimSpace(value)
}

func metricDataType(cfgType string) pmetric.MetricType {
	if cfgType == metricTypeSum {
		return pmetric.MetricTypeSum
	}
	return pmetric.MetricTypeGauge
}

func (p *metricParser) resolve(origin string, elems []string) (resolvedMetric, bool) {
	sub := p.subscriptionFor(origin, elems)
	if sub == nil {
		return resolvedMetric{}, false
	}

	leaf := elems[len(elems)-1]
	if override, ok := sub.Overrides[leaf]; ok {
		return resolvedMetric{MetricConfig: override}, true
	}

	if rm, ok := p.schema.lookup(elems); ok {
		if sub.Default != nil {
			if rm.Unit == "" {
				rm.Unit = sub.Default.Unit
			}
			if len(rm.EnumValues) == 0 {
				rm.EnumValues = sub.Default.EnumValues
				rm.normalizedEnumValues = sub.Default.normalizedEnumValues
			}
		}
		return rm, true
	}

	if sub.Default != nil {
		return resolvedMetric{MetricConfig: *sub.Default}, true
	}

	p.logUnresolved(origin, elems)
	return resolvedMetric{}, false
}

func (p *metricParser) logUnresolved(origin string, elems []string) {
	path := buildMetricName(origin, elems)
	if _, alreadyLogged := p.unresolved.LoadOrStore(path, struct{}{}); alreadyLogged {
		return
	}
	p.logger.Warn("dropping gNMI leaf: no override, schema entry, or default matched it", zap.String("path", path))
}

func (p *metricParser) subscriptionFor(origin string, elems []string) *SubscriptionConfig {
	var best *SubscriptionConfig
	bestLen := -1
	for i := range p.subscriptions {
		sub := &p.subscriptions[i]
		if sub.Origin != "" && sub.Origin != origin {
			continue
		}
		subElems := pathElemNames(sub.Path)
		if len(subElems) > len(elems) {
			continue
		}
		matched := true
		for j, name := range subElems {
			if elems[j] != name {
				matched = false
				break
			}
		}
		if matched && len(subElems) > bestLen {
			best = sub
			bestLen = len(subElems)
		}
	}
	return best
}

func (p *metricParser) emitJSON(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, raw []byte, ts pcommon.Timestamp,
) error {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("invalid JSON payload at %q: %w", buildMetricName(origin, elems), err)
	}
	return p.flatten(b, origin, elems, keys, decoded, ts)
}

func (p *metricParser) flatten(
	b *parseBatch, origin string, elems []string,
	keys map[string]string, value any, ts pcommon.Timestamp,
) error {
	switch v := value.(type) {
	case map[string]any:
		var errs []string
		for name, child := range v {
			child2 := make([]string, len(elems), len(elems)+1)
			copy(child2, elems)
			if err := p.flatten(b, origin, append(child2, name), keys, child, ts); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return errors.New(strings.Join(errs, "; "))
		}
		return nil
	case []any:
		var errs []string
		for i, child := range v {
			if err := p.flatten(b, origin, elems, withIndex(keys, i), child, ts); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return errors.New(strings.Join(errs, "; "))
		}
		return nil
	case float64:
		rm, ok := p.resolve(origin, elems)
		switch {
		case ok && rm.kind == valueKindInt:
			return p.emitInt(b, origin, elems, keys, int64(v), ts)
		case ok && rm.kind == valueKindFloat:
			return p.emitDouble(b, origin, elems, keys, v, ts)
		case v == float64(int64(v)):
			// No schema-derived kind: fall back to the legacy heuristic of
			// treating whole numbers as ints.
			return p.emitInt(b, origin, elems, keys, int64(v), ts)
		default:
			return p.emitDouble(b, origin, elems, keys, v, ts)
		}
	case bool:
		var n int64
		if v {
			n = 1
		}
		return p.emitInt(b, origin, elems, keys, n, ts)
	case string:
		return p.emitInfo(b, origin, elems, keys, v, ts)
	case nil:
		// JSON null carries no value.
		return nil
	default:
		return nil
	}
}

func joinPath(prefix, path *gnmipb.Path) ([]string, map[string]string) {
	type occurrence struct {
		elem  string
		value string
	}
	occurrences := map[string][]occurrence{}
	var elems []string
	for _, p := range []*gnmipb.Path{prefix, path} {
		for _, elem := range p.GetElem() {
			if elem.GetName() != "" {
				elems = append(elems, elem.GetName())
			}
			for k, v := range elem.GetKey() {
				occurrences[k] = append(occurrences[k], occurrence{elem: elem.GetName(), value: v})
			}
		}
	}

	keys := make(map[string]string, len(occurrences))
	for k, occs := range occurrences {
		if len(occs) == 1 {
			keys[k] = occs[0].value
			continue
		}
		for _, occ := range occs {
			name := k
			if occ.elem != "" {
				name = occ.elem + "." + k
			}
			keys[name] = occ.value
		}
	}
	return elems, keys
}

func buildMetricName(origin string, elems []string) string {
	if origin == "" {
		return strings.Join(elems, ".")
	}
	return origin + "." + strings.Join(elems, ".")
}

func pathElemNames(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, "/")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		if idx := strings.IndexByte(part, '['); idx >= 0 {
			part = part[:idx]
		}
		if part != "" {
			names = append(names, part)
		}
	}
	return names
}

func putAttrs(dest pcommon.Map, attrs map[string]string) {
	for k, v := range attrs {
		if v != "" {
			dest.PutStr(k, v)
		}
	}
}
