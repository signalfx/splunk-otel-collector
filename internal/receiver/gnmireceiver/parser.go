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
	"time"

	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	conventions "go.opentelemetry.io/otel/semconv/v1.22.0"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/gnmireceiver/internal/metadata"
)

const (
	infoMetricSuffix = "_info"
	infoValueAttr    = "value"
	// indexAttr identifies an element's position when a JSON array is flattened.
	indexAttr = "index"
)

// metricParser converts gNMI SubscribeResponse messages into OTel metrics.
type metricParser struct {
	endpoint      string
	subscriptions []SubscriptionConfig
}

func newMetricParser(endpoint string, subscriptions []SubscriptionConfig) *metricParser {
	return &metricParser{endpoint: endpoint, subscriptions: subscriptions}
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

	sm := p.newScopeMetrics(metrics)

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

		if err := p.appendUpdate(sm, origin, elems, keys, update.GetVal(), ts); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if sm.Metrics().Len() == 0 {
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
	sm pmetric.ScopeMetrics,
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
		return p.emitUint(sm, origin, elems, keys, v.UintVal, ts)
	case *gnmipb.TypedValue_IntVal:
		p.emitInt(sm, origin, elems, keys, v.IntVal, ts)
	case *gnmipb.TypedValue_FloatVal:
		p.emitDouble(sm, origin, elems, keys, float64(v.FloatVal), ts) //nolint:staticcheck // SA1019: float_val is deprecated but still sent by some targets
	case *gnmipb.TypedValue_DoubleVal:
		p.emitDouble(sm, origin, elems, keys, v.DoubleVal, ts)
	case *gnmipb.TypedValue_BoolVal:
		var n int64
		if v.BoolVal {
			n = 1
		}
		p.emitInt(sm, origin, elems, keys, n, ts)
	case *gnmipb.TypedValue_StringVal:
		p.emitInfo(sm, origin, elems, keys, v.StringVal, ts)
	case *gnmipb.TypedValue_JsonVal:
		return p.emitJSON(sm, origin, elems, keys, v.JsonVal, ts)
	case *gnmipb.TypedValue_JsonIetfVal:
		return p.emitJSON(sm, origin, elems, keys, v.JsonIetfVal, ts)
	case *gnmipb.TypedValue_LeaflistVal:
		var errs []string
		for i, element := range v.LeaflistVal.GetElement() {
			if err := p.appendUpdate(sm, origin, elems, withIndex(keys, i), element, ts); err != nil {
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
	return nil
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
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value uint64, ts pcommon.Timestamp,
) error {
	if value > math.MaxInt64 {
		p.emitDouble(sm, origin, elems, keys, float64(value), ts)
		return fmt.Errorf("value %d for %q exceeds int64 range; emitted as a double and lost precision",
			value, metricName(origin, elems))
	}
	p.emitInt(sm, origin, elems, keys, int64(value), ts)
	return nil
}

func (p *metricParser) emitInt(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value int64, ts pcommon.Timestamp,
) {
	cfg, ok := p.resolve(origin, elems)
	if !ok {
		return
	}
	p.writeInt(sm, origin, elems, keys, cfg, value, ts)
}

func (p *metricParser) emitDouble(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value float64, ts pcommon.Timestamp,
) {
	cfg, ok := p.resolve(origin, elems)
	if !ok {
		return
	}
	p.writeDouble(sm, origin, elems, keys, cfg, value, ts)
}

func (p *metricParser) emitInfo(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value string, ts pcommon.Timestamp,
) {
	cfg, ok := p.resolve(origin, elems)
	if !ok {
		return
	}

	if cfg.Type == metricTypeSum || cfg.Type == metricTypeGauge {
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			p.writeInt(sm, origin, elems, keys, cfg, n, ts)
			return
		}
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			p.writeDouble(sm, origin, elems, keys, cfg, f, ts)
			return
		}
	}

	p.writeInfo(sm, origin, elems, keys, value, ts)
}

func (p *metricParser) writeInt(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, cfg MetricConfig, value int64, ts pcommon.Timestamp,
) {
	dp := p.newNumberDataPoint(sm, origin, elems, cfg)
	dp.SetIntValue(value)
	dp.SetTimestamp(ts)
	putAttrs(dp.Attributes(), keys)
}

func (p *metricParser) writeDouble(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, cfg MetricConfig, value float64, ts pcommon.Timestamp,
) {
	dp := p.newNumberDataPoint(sm, origin, elems, cfg)
	dp.SetDoubleValue(value)
	dp.SetTimestamp(ts)
	putAttrs(dp.Attributes(), keys)
}

func (p *metricParser) writeInfo(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value string, ts pcommon.Timestamp,
) {
	m := sm.Metrics().AppendEmpty()
	m.SetName(metricName(origin, elems) + infoMetricSuffix)
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetIntValue(1)
	dp.SetTimestamp(ts)
	dp.Attributes().PutStr(infoValueAttr, value)
	putAttrs(dp.Attributes(), keys)
}

func (p *metricParser) newNumberDataPoint(
	sm pmetric.ScopeMetrics, origin string, elems []string, cfg MetricConfig,
) pmetric.NumberDataPoint {
	m := sm.Metrics().AppendEmpty()
	m.SetName(metricName(origin, elems))
	m.SetUnit(cfg.Unit)

	if cfg.Type == metricTypeSum {
		sum := m.SetEmptySum()
		sum.SetIsMonotonic(true)
		sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
		return sum.DataPoints().AppendEmpty()
	}
	return m.SetEmptyGauge().DataPoints().AppendEmpty()
}

func (p *metricParser) resolve(origin string, elems []string) (MetricConfig, bool) {
	sub := p.subscriptionFor(origin, elems)
	if sub == nil {
		return MetricConfig{}, false
	}
	leaf := elems[len(elems)-1]
	if cfg, ok := sub.Overrides[leaf]; ok {
		return cfg, true
	}
	if sub.Default != nil {
		return *sub.Default, true
	}
	return MetricConfig{}, false
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
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, raw []byte, ts pcommon.Timestamp,
) error {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("invalid JSON payload at %q: %w", metricName(origin, elems), err)
	}
	p.flatten(sm, origin, elems, keys, decoded, ts)
	return nil
}

func (p *metricParser) flatten(
	sm pmetric.ScopeMetrics, origin string, elems []string,
	keys map[string]string, value any, ts pcommon.Timestamp,
) {
	switch v := value.(type) {
	case map[string]any:
		for name, child := range v {
			child2 := make([]string, len(elems), len(elems)+1)
			copy(child2, elems)
			p.flatten(sm, origin, append(child2, name), keys, child, ts)
		}
	case []any:
		for i, child := range v {
			p.flatten(sm, origin, elems, withIndex(keys, i), child, ts)
		}
	case float64:
		if v == float64(int64(v)) {
			p.emitInt(sm, origin, elems, keys, int64(v), ts)
			return
		}
		p.emitDouble(sm, origin, elems, keys, v, ts)
	case bool:
		var n int64
		if v {
			n = 1
		}
		p.emitInt(sm, origin, elems, keys, n, ts)
	case string:
		p.emitInfo(sm, origin, elems, keys, v, ts)
	case nil:
		// JSON null carries no value.
	}
}

func joinPath(prefix, path *gnmipb.Path) ([]string, map[string]string) {
	keys := map[string]string{}
	var elems []string
	for _, p := range []*gnmipb.Path{prefix, path} {
		for _, elem := range p.GetElem() {
			if elem.GetName() != "" {
				elems = append(elems, elem.GetName())
			}
			for k, v := range elem.GetKey() {
				keys[k] = v
			}
		}
	}
	return elems, keys
}

func metricName(origin string, elems []string) string {
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
