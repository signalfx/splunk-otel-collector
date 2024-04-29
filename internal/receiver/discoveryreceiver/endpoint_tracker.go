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

package discoveryreceiver

import (
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/collector/semconv/v1.22.0"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

const sourcePortAttr = "source.port"

// identifyingAttrKeys are the keys of attributes that are used to identify an entity.
var identifyingAttrKeys = []string{
	semconv.AttributeK8SPodUID,
	semconv.AttributeContainerID,
	semconv.AttributeK8SNodeUID,
	semconv.AttributeHostID,
	sourcePortAttr,
}

var _ observer.Notify = (*notify)(nil)

type endpointTracker struct {
	correlations correlationStore
	config       *Config
	logger       *zap.Logger
	pLogs        chan plog.Logs
	observables  map[component.ID]observer.Observable
	stopCh       chan struct{}
	notifies     []*notify
	// emitInterval defines an interval for emitting entity state events.
	// Potentially can be exposed as a user config option if there is a need.
	emitInterval time.Duration
}

type notify struct {
	observable      observer.Observable
	endpointTracker *endpointTracker
	observerID      component.ID
	id              observer.NotifyID
}

func newEndpointTracker(
	observables map[component.ID]observer.Observable, config *Config, logger *zap.Logger,
	pLogs chan plog.Logs, correlations correlationStore) *endpointTracker {
	return &endpointTracker{
		config:       config,
		observables:  observables,
		logger:       logger,
		pLogs:        pLogs,
		correlations: correlations,
		// 15 minutes is a reasonable default for emitting entity state events given the 1 hour TTL in the inventory
		// service. Potentially we could expose it as a user config, but only if there is a need.
		// Note that we emit entity state events on every entity change. Entities that were changed in the last
		// 15 minutes are not emitted again. So the actual interval of emitting entity state events can be more than 15
		// minutes but always less than 30 minutes.
		emitInterval: 15 * time.Minute,
		stopCh:       make(chan struct{}),
	}
}

func (et *endpointTracker) start() {
	for obs, observable := range et.observables {
		et.logger.Debug("endpointTracker subscribing to observable", zap.Any("observer", obs.String()))
		n := &notify{
			id:              observer.NotifyID(fmt.Sprintf("%p::endpoint_tracker::%s", et, obs.String())),
			observerID:      obs,
			observable:      observable,
			endpointTracker: et,
		}
		et.notifies = append(et.notifies, n)
		go observable.ListAndWatch(n)
	}
	et.correlations.Start()
	go et.startEmitLoop()
}

func (et *endpointTracker) startEmitLoop() {
	timer := time.NewTicker(et.emitInterval)
	for {
		select {
		case corr := <-et.correlations.EmitCh():
			et.emitEntityStateEvents(corr.observerID, []observer.Endpoint{corr.endpoint})
		case <-timer.C:
			for obs := range et.observables {
				et.emitEntityStateEvents(obs, et.correlations.Endpoints(time.Now().Add(-et.emitInterval)))
			}
		case <-et.stopCh:
			timer.Stop()
			return
		}
	}
}

func (et *endpointTracker) stop() {
	for _, n := range et.notifies {
		et.logger.Debug("endpointTracker unsubscribing from observable", zap.Any("observer", n.observerID))
		go n.observable.Unsubscribe(n)
	}
	et.correlations.Stop()
	et.stopCh <- struct{}{}
}

func (et *endpointTracker) emitEntityStateEvents(observerCID component.ID, endpoints []observer.Endpoint) {
	if et.pLogs != nil {
		entityEvents, numFailed, err := entityStateEvents(observerCID, endpoints, et.correlations, time.Now())
		if err != nil {
			et.logger.Warn(fmt.Sprintf("failed converting %v endpoints to entity state events", numFailed), zap.Error(err))
		}
		if entityEvents.Len() > 0 {
			et.pLogs <- entityEvents.ConvertAndMoveToLogs()
		}
	}
}

func (et *endpointTracker) emitEntityDeleteEvents(endpoints []observer.Endpoint) {
	if et.pLogs != nil {
		entityEvents, numFailed, err := entityDeleteEvents(endpoints, time.Now())
		if err != nil {
			et.logger.Warn(fmt.Sprintf("failed converting %v endpoints to entity delete events", numFailed), zap.Error(err))
		}
		if entityEvents.Len() > 0 {
			et.pLogs <- entityEvents.ConvertAndMoveToLogs()
		}
	}
}

func (et *endpointTracker) updateEndpoints(endpoints []observer.Endpoint, observerID component.ID) {
	var matchingEndpoints []observer.Endpoint
	for _, endpoint := range endpoints {
		receiver := et.receiverMatchingEndpoint(endpoint)
		if receiver != discovery.NoType {
			matchingEndpoints = append(matchingEndpoints, endpoint)
			et.correlations.UpdateEndpoint(endpoint, receiver, observerID)
		}
	}
	et.emitEntityStateEvents(observerID, matchingEndpoints)
}

// receiverMatchingEndpoint returns the receiver ID that matches the given endpoint.
// If the endpoint doesn't match exactly one receiver, it returns discovery.NoType.
func (et *endpointTracker) receiverMatchingEndpoint(endpoint observer.Endpoint) component.ID {
	endpointEnv, err := endpoint.Env()
	if err != nil {
		et.logger.Error("failed receiving endpoint environment", zap.String("endpoint", string(endpoint.ID)), zap.Error(err))
		return discovery.NoType
	}
	receivers := et.matchingReceivers(endpointEnv)
	if len(receivers) == 0 {
		et.logger.Debug("endpoint matched no receivers, skipping", zap.String("endpoint", string(endpoint.ID)))
		return discovery.NoType
	}
	if len(receivers) > 1 {
		var receiverNames string
		for _, receiverID := range receivers {
			receiverNames += receiverID.String() + " "
		}
		et.logger.Warn("endpoint matched multiple receivers, skipping", zap.String("endpoint",
			string(endpoint.ID)), zap.String("receivers", receiverNames))
		return discovery.NoType
	}
	return receivers[0]
}

func (et *endpointTracker) matchingReceivers(endpointEnv observer.EndpointEnv) []component.ID {
	var matchingReceivers []component.ID
	for receiverID, receiverCfg := range et.config.Receivers {
		ok, err := receiverCfg.Rule.eval(endpointEnv)
		if err != nil {
			et.logger.Error("failed matching rule", zap.String("rule", receiverCfg.Rule.String()), zap.Error(err))
			continue
		}
		if ok {
			matchingReceivers = append(matchingReceivers, receiverID)
		}
	}
	return matchingReceivers
}

func (n *notify) ID() observer.NotifyID {
	return n.id
}

func (n *notify) OnAdd(added []observer.Endpoint) {
	n.endpointTracker.updateEndpoints(added, n.observerID)
}

func (n *notify) OnRemove(removed []observer.Endpoint) {
	var matchingEndpoints []observer.Endpoint
	for _, endpoint := range removed {
		receiver := n.endpointTracker.receiverMatchingEndpoint(endpoint)
		if receiver != discovery.NoType {
			matchingEndpoints = append(matchingEndpoints, endpoint)
			n.endpointTracker.correlations.MarkStale(endpoint.ID)
		}
	}
	n.endpointTracker.emitEntityDeleteEvents(matchingEndpoints)
}

func (n *notify) OnChange(changed []observer.Endpoint) {
	n.endpointTracker.updateEndpoints(changed, n.observerID)
}

func entityStateEvents(observerID component.ID, endpoints []observer.Endpoint, correlations correlationStore, ts time.Time) (ees experimentalmetricmetadata.EntityEventsSlice, failed int, err error) {
	entityEvents := experimentalmetricmetadata.NewEntityEventsSlice()
	for _, endpoint := range endpoints {
		if endpoint.Details == nil {
			failed++
			err = multierr.Combine(err, fmt.Errorf("endpoint %q has no details", endpoint.ID))
			continue
		}

		entityEvent := entityEvents.AppendEmpty()
		entityEvent.SetTimestamp(pcommon.NewTimestampFromTime(ts))
		entityState := entityEvent.SetEntityState()
		attrs := entityState.Attributes()
		if envAttrs, e := endpointEnvToAttrs(endpoint.Details.Type(), endpoint.Details.Env()); e != nil {
			err = multierr.Combine(err, fmt.Errorf("failed determining attributes for %q: %w", endpoint.ID, e))
			failed++
		} else {
			// this must be the first mutation of attrs since it's destructive
			envAttrs.CopyTo(attrs)
		}
		attrs.PutStr("type", string(endpoint.Details.Type()))
		attrs.PutStr(discovery.EndpointIDAttr, string(endpoint.ID))
		attrs.PutStr("endpoint", endpoint.Target)
		attrs.PutStr(observerNameAttr, observerID.Name())
		attrs.PutStr(observerTypeAttr, observerID.Type().String())
		for k, v := range correlations.Attrs(endpoint.ID) {
			attrs.PutStr(k, v)
		}
		extractIdentifyingAttrs(attrs, entityEvent.ID())
	}
	return entityEvents, failed, err
}

func entityDeleteEvents(endpoints []observer.Endpoint, ts time.Time) (ees experimentalmetricmetadata.EntityEventsSlice, failed int, err error) {
	entityEvents := experimentalmetricmetadata.NewEntityEventsSlice()
	for _, endpoint := range endpoints {
		if endpoint.Details == nil {
			failed++
			err = multierr.Combine(err, fmt.Errorf("endpoint %q has no details", endpoint.ID))
			continue
		}

		entityEvent := entityEvents.AppendEmpty()
		entityEvent.SetTimestamp(pcommon.NewTimestampFromTime(ts))
		entityEvent.SetEntityDelete()
		if envAttrs, e := endpointEnvToAttrs(endpoint.Details.Type(), endpoint.Details.Env()); e != nil {
			err = multierr.Combine(err, fmt.Errorf("failed determining attributes for %q: %w", endpoint.ID, e))
			failed++
		} else {
			extractIdentifyingAttrs(envAttrs, entityEvent.ID())
		}
	}
	return entityEvents, failed, err
}

func endpointEnvToAttrs(endpointType observer.EndpointType, endpointEnv observer.EndpointEnv) (pcommon.Map, error) {
	attrs := pcommon.NewMap()
	for k, v := range endpointEnv {
		switch {
		// labels and annotations for container/node types
		// should result in a ValueMap
		case shouldEmbedMap(endpointType, k):
			if asMap, ok := v.(map[string]string); ok {
				mapVal := attrs.PutEmptyMap(k)
				for item, itemVal := range asMap {
					mapVal.PutStr(item, itemVal)
				}
			} else {
				return attrs, fmt.Errorf("failed parsing %v env attributes", endpointType)
			}
		// pod EndpointEnv is the value of the "pod" field for observer.PortType and should be
		// embedded as ValueMap
		case observer.EndpointType(k) == observer.PodType && endpointType == observer.PortType:
			if podEnv, ok := v.(observer.EndpointEnv); ok {
				podAttrs, e := endpointEnvToAttrs(observer.PodType, podEnv)
				if e != nil {
					return attrs, fmt.Errorf("failed parsing %v pod attributes ", endpointType)
				}
				podAttrs.Range(func(k string, v pcommon.Value) bool {
					v.CopyTo(attrs.PutEmpty(k))
					return true
				})
			} else {
				return attrs, fmt.Errorf("failed parsing %v pod env %#v", endpointType, v)
			}

		// rename keys according to the OTel Semantic Conventions
		case k == "container_id":
			attrs.PutEmpty(semconv.AttributeContainerID).FromRaw(v)
		case k == "port":
			attrs.PutEmpty(sourcePortAttr).FromRaw(v)
		case endpointType == observer.PodType:
			if k == "namespace" {
				attrs.PutEmpty(semconv.AttributeK8SNamespaceName).FromRaw(v)
			} else {
				attrs.PutEmpty("k8s.pod." + k).FromRaw(v)
			}
		case endpointType == observer.K8sNodeType:
			switch k {
			case "name":
				attrs.PutEmpty(semconv.AttributeK8SNodeName).FromRaw(v)
			case "uid":
				attrs.PutEmpty(semconv.AttributeK8SNodeUID).FromRaw(v)
			default:
				attrs.PutEmpty(k).FromRaw(v)
			}
		default:
			switch vVal := v.(type) {
			case uint16:
				attrs.PutInt(k, int64(vVal))
			case bool:
				attrs.PutBool(k, vVal)
			default:
				attrs.PutStr(k, fmt.Sprintf("%v", v))
			}
		}
	}
	return attrs, nil
}

func shouldEmbedMap(endpointType observer.EndpointType, k string) bool {
	return (k == "annotations" && (endpointType == observer.PodType ||
		endpointType == observer.K8sNodeType)) ||
		(k == "labels" && (endpointType == observer.PodType ||
			endpointType == observer.ContainerType ||
			endpointType == observer.K8sNodeType))
}

func extractIdentifyingAttrs(from pcommon.Map, to pcommon.Map) {
	for _, k := range identifyingAttrKeys {
		if v, ok := from.Get(k); ok {
			v.CopyTo(to.PutEmpty(k))
			from.Remove(k)
		}
	}
}
