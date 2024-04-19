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
	"sync"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

// correlation is a grouping of an endpoint, the observing
// observer, and a receiver, if any. It's used to unify
// emitted log record content from evaluated status sources
// such that no individual source lacks information otherwise
// only available to others. At this time this means that
// metrics and component logs will have access to a receiver's
// observing observer via the Notify event where not available
// through another means.
type correlation struct {
	lastState   endpointState
	lastUpdated time.Time
	endpoint    observer.Endpoint
	receiverID  component.ID
	observerID  component.ID
}

// correlationStore provides a centralized interface for up-to-date correlations
// and receiver attributes as a message passing mechanism by observed components.
// It manages a reaping loop to prevent stale endpoint buildup over time.
type correlationStore interface {
	UpdateEndpoint(endpoint observer.Endpoint, receiverID component.ID, state endpointState, observerID component.ID)
	GetOrCreate(receiverID component.ID, endpointID observer.EndpointID) correlation
	Attrs(receiverID component.ID) map[string]string
	UpdateAttrs(receiverID component.ID, attrs map[string]string)
	Endpoints(updatedBefore time.Time) []observer.Endpoint
	// Start the reaping loop to prevent unnecessary endpoint buildup
	Start()
	// Stop the reaping loop
	Stop()
}

// store is a collection of mappings used as an instantaneous record of
// 1. endpoints to their associated correlations
// 2. receivers to their endpoint-agnostic Attrs used as a message
// passing mechanism (currently just for embedded config values).
// This way the pre-created receiver instances can have attrs
// before they are created via endpoint.
type store struct {
	logger *zap.Logger
	// correlations is a ~synchronized map[endpointID]*corr
	correlations  *sync.Map
	endpointLocks *keyLock
	receiverAttrs *sync.Map
	receiverLocks *keyLock
	// sentinel for terminating reaper loop
	sentinel     chan struct{}
	reapInterval time.Duration
	ttl          time.Duration
}

func newCorrelationStore(logger *zap.Logger, ttl time.Duration) correlationStore {
	return &store{
		logger:        logger,
		correlations:  &sync.Map{},
		endpointLocks: newKeyLock(),
		receiverAttrs: &sync.Map{},
		receiverLocks: newKeyLock(),
		reapInterval:  30 * time.Second,
		ttl:           ttl,
		sentinel:      make(chan struct{}, 1),
	}
}

// UpdateEndpoint updates or creates correlation timestamps and states by endpoint.ID and receiverID.
func (s *store) UpdateEndpoint(endpoint observer.Endpoint, receiverID component.ID, state endpointState, observerID component.ID) {
	endpointUnlock := s.endpointLocks.Lock(endpoint.ID)
	defer endpointUnlock()
	c, ok := s.correlations.LoadOrStore(endpoint.ID, &correlation{
		endpoint:    endpoint,
		receiverID:  receiverID,
		observerID:  observerID,
		lastState:   state,
		lastUpdated: time.Now(),
	})
	if ok {
		corr := c.(*correlation)
		corr.receiverID = receiverID
		corr.endpoint = endpoint
		// set here for unlikely out of order GetOrCreate eventual consistency
		corr.observerID = observerID
		corr.lastUpdated = time.Now()
		corr.lastState = state
	}
}

// Endpoints returns all active endpoints that have not been updated since the provided time.
func (s *store) Endpoints(updatedBefore time.Time) []observer.Endpoint {
	var endpoints []observer.Endpoint
	s.correlations.Range(func(eID, c any) bool {
		endpointID := eID.(observer.EndpointID)
		endpointUnlock := s.endpointLocks.Lock(endpointID)
		defer endpointUnlock()
		corr := c.(*correlation)
		if corr.lastState != removedState && corr.lastUpdated.Before(updatedBefore) {
			endpoints = append(endpoints, c.(*correlation).endpoint)
		}
		return true
	})
	return endpoints
}

// GetOrCreate returns an existing receiver/endpoint correlation or creates a new one.
func (s *store) GetOrCreate(receiverID component.ID, endpointID observer.EndpointID) correlation {
	endpointUnlock := s.endpointLocks.Lock(endpointID)
	defer endpointUnlock()
	c, ok := s.correlations.Load(endpointID)
	if ok {
		return *(c.(*correlation))
	}
	// The observer has yet to emit an endpoint event and this could be an invalid collector state.
	corr := correlation{
		receiverID: receiverID,
		observerID: discovery.NoType,
	}
	s.correlations.Store(endpointID, &corr)
	return corr
}

func (s *store) Attrs(receiverID component.ID) map[string]string {
	defer s.receiverLocks.Lock(receiverID)()
	rInfo, _ := s.receiverAttrs.LoadOrStore(receiverID, map[string]string{})
	receiverInfo := rInfo.(map[string]string)
	cp := map[string]string{}
	for k, v := range receiverInfo {
		cp[k] = v
	}
	return cp
}

func (s *store) UpdateAttrs(receiverID component.ID, attrs map[string]string) {
	defer s.receiverLocks.Lock(receiverID)()
	rAttrs, _ := s.receiverAttrs.LoadOrStore(receiverID, map[string]string{})
	receiverAttrs := rAttrs.(map[string]string)
	for k, v := range attrs {
		receiverAttrs[k] = v
	}
	s.receiverAttrs.Store(receiverID, receiverAttrs)
}

func (s *store) Start() {
	go func() {
		timer := time.NewTicker(s.reapInterval)
		for {
			select {
			case <-timer.C:
				s.reap()
			case <-s.sentinel:
				timer.Stop()
				return
			}
		}
	}()
}

func (s *store) Stop() {
	s.sentinel <- struct{}{}
}

// reap will remove all removed endpoints whose last update is past the ttl
func (s *store) reap() {
	s.correlations.Range(func(eID, c any) bool {
		endpointID := eID.(observer.EndpointID)
		endpointUnlock := s.endpointLocks.Lock(endpointID)
		defer endpointUnlock()
		corr := c.(*correlation)
		if corr.lastState == removedState && time.Since(corr.lastUpdated) > s.ttl {
			s.correlations.Delete(endpointID)
		}
		return true
	})
}

// keyLock is a map of locks for an associated Map to be locked
// by its keys for a longer duration than the provided atomic ops.
type keyLock struct {
	*sync.Map
}

func newKeyLock() *keyLock {
	return &keyLock{&sync.Map{}}
}

func (kl *keyLock) Lock(key any) (unlock func()) {
	mtx, _ := kl.Map.LoadOrStore(key, &sync.Mutex{})
	mutex := mtx.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
}
