package observers

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/services"
)

// ServiceDiffer will run the DiscoveryFn every IntervalSeconds and report
// any new or removed services to the provided Callbacks.
type ServiceDiffer struct {
	DiscoveryFn     func() []services.Endpoint
	IntervalSeconds int
	Callbacks       *ServiceCallbacks
	serviceSet      map[services.ID]services.Endpoint
	stop            chan struct{}
}

// Start polling the DiscoveryFn on a regular interval
func (sd *ServiceDiffer) Start() {
	sd.serviceSet = make(map[services.ID]services.Endpoint)
	sd.stop = make(chan struct{})

	ticker := time.NewTicker(time.Duration(sd.IntervalSeconds) * time.Second)

	// Do discovery immediately so that services can be monitored ASAP
	sd.runDiscovery()

	go func() {
		for {
			select {
			case <-sd.stop:
				close(sd.stop)
				sd.stop = nil
				ticker.Stop()
				return
			case <-ticker.C:
				sd.runDiscovery()
			}
		}
	}()
}

func (sd *ServiceDiffer) runDiscovery() {
	latestServices := sd.DiscoveryFn()

	// Assume a service is inactive until told otherwise by the
	// discovery function
	activeSet := map[services.ID]bool{}
	for sid := range sd.serviceSet {
		activeSet[sid] = false
	}

	for i := range latestServices {
		service := latestServices[i]
		_, seen := sd.serviceSet[service.Core().ID]
		if !seen {
			log.WithFields(log.Fields{
				"serviceID": service.Core().ID,
			}).Debug("ServiceDiffer: adding service")

			sd.Callbacks.Added(service)
		}
		sd.serviceSet[service.Core().ID] = service
		activeSet[service.Core().ID] = true
	}

	// Remove any that are no longer reported by discovery
	for sid, active := range activeSet {
		if !active {
			sd.Callbacks.Removed(sd.serviceSet[sid])
			delete(sd.serviceSet, sid)
		}
	}
}

// Stop polling the DiscoveryFn
func (sd *ServiceDiffer) Stop() {
	if sd.stop != nil {
		sd.stop <- struct{}{}
	}
}
