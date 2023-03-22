package observers

import (
	"fmt"
	"strings"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/sfxclient"
)

// DiagnosticText outputs human-readable text about the active observers.
func (om *ObserverManager) DiagnosticText() string {
	observerTypes := make([]string, len(om.observers))
	for i := range om.observers {
		observerTypes[i] = om.observers[i]._type
	}
	return fmt.Sprintf("Observers active:                 %s", strings.Join(observerTypes, ", "))
}

// InternalMetrics returns a list of datapoints relevant to the internal status
// of Observers
func (om *ObserverManager) InternalMetrics() []*datapoint.Datapoint {
	return []*datapoint.Datapoint{
		sfxclient.Gauge("sfxagent.active_observers", nil, int64(len(om.observers))),
	}
}
