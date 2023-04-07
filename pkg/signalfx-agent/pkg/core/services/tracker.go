package services

import (
	"errors"
	"sync"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
	"github.com/sirupsen/logrus"
)

type EndpointsByID map[ID]Endpoint

func (e EndpointsByID) First() (Endpoint, error) {
	for id := range e {
		return e[id], nil
	}
	return nil, errors.New("no endpoints present")
}

func (e EndpointsByID) AddEndpoint(endpoint Endpoint) EndpointsByID {
	id := endpoint.Core().ID
	if e == nil {
		return EndpointsByID{
			id: endpoint,
		}
	}
	e[id] = endpoint
	return e
}

func (e EndpointsByID) RemoveEndpoint(endpoint Endpoint) {
	if e == nil {
		return
	}
	delete(e, endpoint.Core().ID)
}

func (e EndpointsByID) AsSlice() []Endpoint {
	if e == nil {
		return nil
	}

	var out []Endpoint
	for _, endpoint := range e {
		out = append(out, endpoint)
	}
	return out
}

// EndpointHostTracker is used to maintain the relationship between an
// endpoint's IP address (host) and the endpoint(s) that pertain to it.
type EndpointHostTracker struct {
	sync.RWMutex
	hostToEndpoints map[string]EndpointsByID
}

func NewEndpointHostTracker() *EndpointHostTracker {
	return &EndpointHostTracker{
		hostToEndpoints: make(map[string]EndpointsByID),
	}
}

func (et *EndpointHostTracker) EndpointAdded(endpoint Endpoint) {
	et.Lock()
	defer et.Unlock()

	host := endpoint.Core().Host
	if host == "" {
		return
	}

	logrus.Debugf("Mapping host %s to endpoint %v", host, endpoint)
	et.hostToEndpoints[host] = et.hostToEndpoints[host].AddEndpoint(endpoint)
}

func (et *EndpointHostTracker) EndpointRemoved(endpoint Endpoint) {
	et.Lock()
	defer et.Unlock()

	host := endpoint.Core().Host
	if host == "" {
		return
	}

	et.hostToEndpoints[host].RemoveEndpoint(endpoint)
}

func (et *EndpointHostTracker) GetByHost(host string) []Endpoint {
	et.RLock()
	defer et.RUnlock()

	return et.hostToEndpoints[host].AsSlice()
}

func (et *EndpointHostTracker) InternalMetrics() []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		sfxclient.Cumulative("sfxagent.endpoint_host_tracker_size", nil, int64(len(et.hostToEndpoints))),
	}
}
