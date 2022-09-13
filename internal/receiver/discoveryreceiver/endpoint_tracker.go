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
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"
	"go.uber.org/zap"
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
	logger       *zap.Logger
	pLogs        chan plog.Logs
	observables  map[config.ComponentID]observer.Observable
	correlations correlationStore
	notifies     []*notify
	logEndpoints bool
}

type notify struct {
	observable      observer.Observable
	endpointTracker *endpointTracker
	observerID      config.ComponentID
	id              observer.NotifyID
}

func newEndpointTracker(
	observables map[config.ComponentID]observer.Observable, config *Config, logger *zap.Logger,
	pLogs chan plog.Logs, correlations correlationStore) *endpointTracker {
	return &endpointTracker{
		logEndpoints: config.LogEndpoints,
		observables:  observables,
		logger:       logger,
		pLogs:        pLogs,
		correlations: correlations,
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
}

func (et *endpointTracker) stop() {
	for _, n := range et.notifies {
		et.logger.Debug("endpointTracker unsubscribing from observable", zap.Any("observer", n.observerID))
		go n.observable.Unsubscribe(n)
	}
	et.correlations.Stop()
}

func (et *endpointTracker) emitEndpointLogs(observerCID config.ComponentID, eventType endpointState, endpoints []observer.Endpoint, received time.Time) {
	if et.logEndpoints && et.pLogs != nil {
		pLogs, numFailed, err := endpointToPLogs(observerCID, fmt.Sprintf("endpoint.%s", eventType), endpoints, received)
		if err != nil {
			et.logger.Warn(fmt.Sprintf("failed converting %v endpoints to log records", numFailed), zap.Error(err))
		}
		if pLogs.LogRecordCount() > 0 {
			et.pLogs <- pLogs
		}
	}
}

func (et *endpointTracker) updateEndpoints(endpoints []observer.Endpoint, state endpointState, observerID config.ComponentID) {
	for _, endpoint := range endpoints {
		et.correlations.UpdateEndpoint(endpoint, state, observerID)
	}
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

func endpointToPLogs(observerID config.ComponentID, eventType string, endpoints []observer.Endpoint, received time.Time) (pLogs plog.Logs, failed int, err error) {
	pLogs = plog.NewLogs()
	rlog := pLogs.ResourceLogs().AppendEmpty()
	rAttrs := rlog.Resource().Attributes()
	rAttrs.UpsertString(eventTypeAttr, eventType)
	rAttrs.UpsertString(observerNameAttr, observerID.Name())
	rAttrs.UpsertString(observerTypeAttr, string(observerID.Type()))
	sl := rlog.ScopeLogs().AppendEmpty()
	for _, endpoint := range endpoints {
		logRecord := sl.LogRecords().AppendEmpty()
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(received))
		logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		logRecord.SetSeverityText("info")
		attrs := logRecord.Attributes()
		if endpoint.Details != nil {
			logRecord.Body().SetStringVal(fmt.Sprintf("%s %s endpoint %s", eventType, endpoint.Details.Type(), endpoint.ID))
			if envAttrs, e := endpointEnvToAttrs(endpoint.Details.Type(), endpoint.Details.Env()); e != nil {
				err = multierr.Combine(err, fmt.Errorf("failed determining attributes for %q: %w", endpoint.ID, e))
				failed++
			} else {
				// this must be the first mutation of attrs since it's destructive
				envAttrs.CopyTo(attrs)
			}
			attrs.UpsertString("type", string(endpoint.Details.Type()))
		} else {
			logRecord.Body().SetStringVal(fmt.Sprintf("%s endpoint %s", eventType, endpoint.ID))
		}
		attrs.UpsertString("endpoint", endpoint.Target)
		attrs.UpsertString("id", string(endpoint.ID))

		// sorted log record attributes for determinism
		attrs.Sort()
	}
	return
}

func endpointEnvToAttrs(endpointType observer.EndpointType, endpointEnv observer.EndpointEnv) (pcommon.Map, error) {
	attrs := pcommon.NewMap()
	for k, v := range endpointEnv {
		switch {
		// labels and annotations for container/node types
		// should result in a ValueMap
		case shouldEmbedMap(endpointType, k):
			if asMap, ok := v.(map[string]string); ok {
				mapVal := attrs.UpsertEmptyMap(k)
				for item, itemVal := range asMap {
					mapVal.UpsertString(item, itemVal)
				}
				mapVal.Sort()
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
				podAttrs.CopyTo(attrs.UpsertEmptyMap(k))
			} else {
				return attrs, fmt.Errorf("failed parsing %v pod env %#v", endpointType, v)
			}
		default:
			switch vVal := v.(type) {
			case uint16:
				attrs.UpsertInt(k, int64(vVal))
			case bool:
				attrs.UpsertBool(k, vVal)
			default:
				attrs.UpsertString(k, fmt.Sprintf("%v", v))
			}
		}
	}
	attrs.Sort()
	return attrs, nil
}

func shouldEmbedMap(endpointType observer.EndpointType, k string) bool {
	return (k == "annotations" && (endpointType == observer.PodType ||
		endpointType == observer.K8sNodeType)) ||
		(k == "labels" && (endpointType == observer.PodType ||
			endpointType == observer.ContainerType ||
			endpointType == observer.K8sNodeType))
}
