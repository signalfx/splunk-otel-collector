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
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

type endpointState string

const (
	addedState   endpointState = "added"
	removedState endpointState = "removed"
	changedState endpointState = "changed"
)

var (
	_ observer.Notify = (*notify)(nil)
)

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
		case <-timer.C:
			for obs := range et.observables {
				activeEndpoints := et.correlations.Endpoints(time.Now().Add(-et.emitInterval))
				// changedState just means that we want to report the current state of the endpoint.
				et.emitEndpointLogs(obs, changedState, activeEndpoints, time.Now())
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

func (et *endpointTracker) emitEndpointLogs(observerCID component.ID, eventType endpointState, endpoints []observer.Endpoint, received time.Time) {
	if et.config.LogEndpoints && et.pLogs != nil {
		pLogs, numFailed, err := endpointToPLogs(observerCID, eventType, endpoints, received)
		if err != nil {
			et.logger.Warn(fmt.Sprintf("failed converting %v endpoints to log records", numFailed), zap.Error(err))
		}
		if pLogs.LogRecordCount() > 0 {
			et.pLogs <- pLogs
		}
	}
}

func (et *endpointTracker) updateEndpoints(endpoints []observer.Endpoint, state endpointState, observerID component.ID) {
	for _, endpoint := range endpoints {
		endpointEnv, err := endpoint.Env()
		if err != nil {
			et.logger.Error("failed receiving endpoint environment", zap.String("endpoint", string(endpoint.ID)), zap.Error(err))
			continue
		}
		receivers := et.matchingReceivers(endpointEnv)
		if len(receivers) == 0 {
			et.logger.Debug("endpoint matched no receivers, skipping", zap.String("endpoint", string(endpoint.ID)))
			continue
		}
		if len(receivers) > 1 {
			var receiverNames string
			for _, receiverID := range receivers {
				receiverNames += receiverID.String() + " "
			}
			et.logger.Warn("endpoint matched multiple receivers, skipping", zap.String("endpoint",
				string(endpoint.ID)), zap.String("receivers", receiverNames))
			continue
		}
		et.correlations.UpdateEndpoint(endpoint, receivers[0], state, observerID)
	}
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
	n.endpointTracker.emitEndpointLogs(n.observerID, addedState, added, time.Now())
	n.endpointTracker.updateEndpoints(added, addedState, n.observerID)
}

func (n *notify) OnRemove(removed []observer.Endpoint) {
	n.endpointTracker.emitEndpointLogs(n.observerID, removedState, removed, time.Now())
	n.endpointTracker.updateEndpoints(removed, removedState, n.observerID)
}

func (n *notify) OnChange(changed []observer.Endpoint) {
	n.endpointTracker.emitEndpointLogs(n.observerID, changedState, changed, time.Now())
	n.endpointTracker.updateEndpoints(changed, changedState, n.observerID)
}

func endpointToPLogs(observerID component.ID, eventType endpointState, endpoints []observer.Endpoint, received time.Time) (pLogs plog.Logs, failed int, err error) {
	entityEvents := experimentalmetricmetadata.NewEntityEventsSlice()
	for _, endpoint := range endpoints {
		entityEvent := entityEvents.AppendEmpty()
		entityEvent.SetTimestamp(pcommon.NewTimestampFromTime(received))
		entityEvent.ID().PutStr(discovery.EndpointIDAttr, string(endpoint.ID))
		if eventType == removedState {
			entityEvent.SetEntityDelete()
		} else {
			entityState := entityEvent.SetEntityState()
			attrs := entityState.Attributes()
			if endpoint.Details != nil {
				if envAttrs, e := endpointEnvToAttrs(endpoint.Details.Type(), endpoint.Details.Env()); e != nil {
					err = multierr.Combine(err, fmt.Errorf("failed determining attributes for %q: %w", endpoint.ID, e))
					failed++
				} else {
					// this must be the first mutation of attrs since it's destructive
					envAttrs.CopyTo(attrs)
				}
				attrs.PutStr("type", string(endpoint.Details.Type()))
			}
			attrs.PutStr("endpoint", endpoint.Target)
			attrs.PutStr(observerNameAttr, observerID.Name())
			attrs.PutStr(observerTypeAttr, observerID.Type().String())
		}
	}
	return entityEvents.ConvertAndMoveToLogs(), failed, err
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
				podAttrs.CopyTo(attrs.PutEmptyMap(k))
			} else {
				return attrs, fmt.Errorf("failed parsing %v pod env %#v", endpointType, v)
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
