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
	"regexp"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	conventions "go.opentelemetry.io/otel/semconv/v1.22.0"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

const (
	entityType      = "service"
	sourcePortAttr  = "source.port"
	serviceTypeAttr = "service.type"
)

// identifyingAttrKeys are the keys of attributes that are used to identify an entity.
var identifyingAttrKeys = []string{
	serviceTypeAttr,
	string(conventions.ServiceNameKey),
	string(conventions.K8SPodUIDKey),
	string(conventions.ContainerIDKey),
	string(conventions.K8SNodeUIDKey),
	string(conventions.HostIDKey),
	sourcePortAttr,
}

// k8sPodRegexp is a regular expression to extract the owner object name from a pod name.
// Built based on https://github.com/kubernetes/apimachinery/blob/d5c9711b77ee5a0dde0fef41c9ca86a67f5ddb4e/pkg/util/rand/rand.go#L92-L127
var k8sPodRegexp = regexp.MustCompile(`^(.+?)-(?:(?:[0-9bcdf]+-)?[bcdfghjklmnpqrstvwxz2456789]{5}|[0-9]+)$`)

var _ observer.Notify = (*notify)(nil)

type endpointTracker struct {
	correlations *correlationStore
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
	pLogs chan plog.Logs, correlations *correlationStore) *endpointTracker {
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
		case corr := <-et.correlations.emitCh:
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
		entityEvents, numFailed, err := entityEvents(observerCID, endpoints, et.correlations, time.Now(), experimentalmetricmetadata.EventTypeState)
		if err != nil {
			et.logger.Warn(fmt.Sprintf("failed converting %v endpoints to entity state events", numFailed), zap.Error(err))
		}
		if entityEvents.Len() > 0 {
			et.pLogs <- entityEvents.ConvertAndMoveToLogs()
		}
	}
}

func (et *endpointTracker) emitEntityDeleteEvents(observerCID component.ID, endpoints []observer.Endpoint) {
	if et.pLogs != nil {
		entityEvents, numFailed, err := entityEvents(observerCID, endpoints, et.correlations, time.Now(),
			experimentalmetricmetadata.EventTypeDelete)
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
	n.endpointTracker.emitEntityDeleteEvents(n.observerID, matchingEndpoints)
}

func (n *notify) OnChange(changed []observer.Endpoint) {
	n.endpointTracker.updateEndpoints(changed, n.observerID)
}

// entityEvents converts observer endpoints to entity state events excluding those
// that don't have a discovery status attribute yet.
func entityEvents(observerID component.ID, endpoints []observer.Endpoint, correlations *correlationStore,
	ts time.Time, eventType experimentalmetricmetadata.EventType) (ees experimentalmetricmetadata.EntityEventsSlice, failed int, err error) {
	events := experimentalmetricmetadata.NewEntityEventsSlice()
	for _, endpoint := range endpoints {
		if endpoint.Details == nil {
			failed++
			err = multierr.Combine(err, fmt.Errorf("endpoint %q has no details", endpoint.ID))
			continue
		}

		endpointAttrs := correlations.Attrs(endpoint.ID)
		if _, ok := endpointAttrs[discovery.StatusAttr]; !ok {
			// If the endpoint doesn't have a status attribute, it's not ready to be emitted.
			continue
		}

		attrs, e := endpointEnvToAttrs(endpoint.Details.Type(), endpoint.Details.Env())
		if e != nil {
			err = multierr.Combine(err, fmt.Errorf("failed determining attributes for %q: %w", endpoint.ID, e))
			failed++
			continue
		}
		attrs.PutStr("type", string(endpoint.Details.Type()))
		attrs.PutStr(discovery.EndpointIDAttr, string(endpoint.ID))
		attrs.PutStr("endpoint", endpoint.Target)
		attrs.PutStr(observerNameAttr, observerID.Name())
		attrs.PutStr(observerTypeAttr, observerID.Type().String())
		for k, v := range endpointAttrs {
			attrs.PutStr(k, v)
		}
		attrs.PutStr(serviceTypeAttr, deduceServiceType(attrs))
		attrs.PutStr(string(conventions.ServiceNameKey), deduceServiceName(attrs))

		event := events.AppendEmpty()
		event.SetTimestamp(pcommon.NewTimestampFromTime(ts))
		extractIdentifyingAttrs(attrs, event.ID())
		switch eventType {
		case experimentalmetricmetadata.EventTypeState:
			entityState := event.SetEntityState()
			entityState.SetEntityType(entityType)
			attrs.MoveTo(entityState.Attributes())
		case experimentalmetricmetadata.EventTypeDelete:
			deleteEvent := event.SetEntityDelete()
			deleteEvent.SetEntityType(entityType)
		}
	}
	return events, failed, err
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
			attrs.PutEmpty(string(conventions.ContainerIDKey)).FromRaw(v)
		case k == "port":
			attrs.PutEmpty(sourcePortAttr).FromRaw(v)
		case endpointType == observer.PodType:
			if k == "namespace" {
				attrs.PutEmpty(string(conventions.K8SNamespaceNameKey)).FromRaw(v)
			} else {
				attrs.PutEmpty("k8s.pod." + k).FromRaw(v)
			}
		case endpointType == observer.K8sNodeType:
			switch k {
			case "name":
				attrs.PutEmpty(string(conventions.K8SNodeNameKey)).FromRaw(v)
			case "uid":
				attrs.PutEmpty(string(conventions.K8SNodeUIDKey)).FromRaw(v)
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

func deduceServiceName(attrs pcommon.Map) string {
	if val, ok := attrs.Get(string(conventions.ServiceNameKey)); ok {
		return val.AsString()
	}
	if labels, labelsFound := attrs.Get("labels"); labelsFound && labels.Type() == pcommon.ValueTypeMap {
		if val, ok := labels.Map().Get("app.kubernetes.io/name"); ok {
			return val.AsString()
		}
		if val, ok := labels.Map().Get("app"); ok {
			return val.AsString()
		}
	}
	// TODO: Update the observer upstream to set the deployment/statefulset name in addition to the pod name,
	//       so we don't have to extract it from the pod name.
	if podName, ok := attrs.Get("k8s.pod.name"); ok {
		matches := k8sPodRegexp.FindStringSubmatch(podName.AsString())
		if len(matches) > 1 {
			return matches[1]
		}
	}
	if val, ok := attrs.Get("name"); ok {
		return val.AsString()
	}
	if val, ok := attrs.Get("process_name"); ok {
		return val.AsString()
	}
	return "unknown"
}

func deduceServiceType(attrs pcommon.Map) string {
	if val, ok := attrs.Get(serviceTypeAttr); ok {
		return val.AsString()
	}
	if val, ok := attrs.Get(discovery.ReceiverTypeAttr); ok {
		return val.AsString()
	}
	return "unknown"
}
