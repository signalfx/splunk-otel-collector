// Copyright  Splunk, Inc.
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
