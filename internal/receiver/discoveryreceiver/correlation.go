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
	"go.opentelemetry.io/collector/config"
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
	receiverID  config.ComponentID
	observerID  config.ComponentID
}

// correlationStore provides a centralized interface for up-to-date correlations
// and receiver attributes as a message passing mechanism by observed components.
// It manages a reaping loop to prevent stale endpoint buildup over time.
type correlationStore interface {
	UpdateEndpoint(endpoint observer.Endpoint, state endpointState, observerID config.ComponentID)
	GetOrCreate(receiverID config.ComponentID, endpointID observer.EndpointID) correlation
	Attrs(receiverID config.ComponentID) map[string]string
	UpdateAttrs(receiverID config.ComponentID, attrs map[string]string)
	// Start the reaping loop to prevent unnecessary endpoint buildup
	Start()
	// Stop the reaping loop
	Stop()
}

// store is a collection of mappings used as an instantaneous record of
// 1. endpoints to their associated receivers->correlations
// 2. receivers to their endpoint-agnostic Attrs used as a message
// passing mechanism (currently just for embedded config values).
// This way the pre-created receiver instances can have attrs
// before they are created via endpoint.
type store struct {
	logger *zap.Logger
	// correlations is a ~synchronized map[endpointID]map[receiverID]*corr
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

// UpdateEndpoint will update all existing correlation timestamps and states by endpoint.ID, or
// creates a new no-type ~singleton w/ the initial correlation info for later use in correlation creation.
func (s *store) UpdateEndpoint(endpoint observer.Endpoint, state endpointState, observerID config.ComponentID) {
	defer s.endpointLocks.Lock(endpoint.ID)()
	rMap, ok := s.correlations.LoadOrStore(endpoint.ID, &sync.Map{})
	receiverMap := rMap.(*sync.Map)
	if !ok {
		receiverMap.Store(discovery.NoType, &correlation{
			endpoint:    endpoint,
			observerID:  observerID,
			lastState:   state,
			lastUpdated: time.Now(),
		})
		// we've set NoType correlation, which is all we can do
		return
	}
	receiverMap.Range(func(_, c any) bool {
		corr := c.(*correlation)
		corr.endpoint = endpoint
		// set here for unlikely out of order GetOrCreate eventual consistency
		corr.observerID = observerID
		corr.lastUpdated = time.Now()
		corr.lastState = state
		return true
	})
}

// GetOrCreate returns an existing receiver/endpoint correlation or creates a new one
// based on the no-type ~singleton for the last endpoint update event.
func (s *store) GetOrCreate(receiverID config.ComponentID, endpointID observer.EndpointID) correlation {
	endpointUnlock := s.endpointLocks.Lock(endpointID)
	rMap, ok := s.correlations.LoadOrStore(endpointID, &sync.Map{})
	receiverMap := rMap.(*sync.Map)
	if !ok {
		// this zero value correlation suggests the observer has yet to emit an endpoint event
		// and this could be an invalid collector state. Likely this flow results from delayed
		// event handling and the correlation is eventually consistent in UpdateEndpoint.
		receiverMap.Store(discovery.NoType, &correlation{})
	}
	defer s.receiverLocks.Lock(receiverID)()
	endpointUnlock()
	c, ok := receiverMap.Load(receiverID)
	if ok {
		return *(c.(*correlation))
	}
	var noTypeCorrelation *correlation
	// disregard ok since previous LoadOrStore handling guarantees existence
	ntCorr, _ := receiverMap.Load(discovery.NoType)
	noTypeCorrelation = ntCorr.(*correlation)
	cpCorr := *noTypeCorrelation
	corr := &cpCorr
	corr.receiverID = receiverID
	receiverMap.Store(receiverID, corr)
	return *corr
}

func (s *store) Attrs(receiverID config.ComponentID) map[string]string {
	defer s.receiverLocks.Lock(receiverID)()
	rInfo, _ := s.receiverAttrs.LoadOrStore(receiverID, map[string]string{})
	receiverInfo := rInfo.(map[string]string)
	cp := map[string]string{}
	for k, v := range receiverInfo {
		cp[k] = v
	}
	return cp
}

func (s *store) UpdateAttrs(receiverID config.ComponentID, attrs map[string]string) {
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
	s.correlations.Range(func(eID, rMap any) bool {
		endpointID := eID.(observer.EndpointID)
		defer s.endpointLocks.Lock(endpointID)()
		receiverMap := rMap.(*sync.Map)
		if c, ok := receiverMap.Load(discovery.NoType); ok {
			corr := c.(*correlation)
			if corr.lastState == removedState &&
				time.Since(corr.lastUpdated) > s.ttl {
				s.correlations.Delete(endpointID)
			}
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
