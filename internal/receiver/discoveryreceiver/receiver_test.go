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
	"context"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/extension"
	otelcolreceiver "go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestNewDiscoveryReceiver(t *testing.T) {
	rcs := otelcolreceiver.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger:         zap.NewNop(),
			MeterProvider:  noop.NewMeterProvider(),
			TracerProvider: trace.NewNoopTracerProvider(),
		},
	}
	cfg := &Config{}
	receiver, err := newDiscoveryReceiver(rcs, cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, receiver)

	// out of order shutdown
	require.NoError(t, receiver.Shutdown(context.Background()))
}

func TestObservablesFromHost(t *testing.T) {
	nopObsID := component.NewID("nop_observer")
	nopObs := &nopObserver{}
	nopObsIDWithName := component.NewIDWithName("nop_observer", "with_name")
	nopObsWithName := &nopObserver{}
	nopObsvbleID := component.NewID("nop_observable")
	nopObsvble := &nopObservable{}
	nopObsvbleIDWithName := component.NewIDWithName("nop_observable", "with_name")
	nopObsvbleWithName := &nopObservable{}

	for _, tt := range []struct {
		name                string
		extensions          map[component.ID]extension.Extension
		expectedObservables map[component.ID]observer.Observable
		expectedError       string
		watchObservers      []component.ID
	}{
		{name: "mixed non-observables ids",
			extensions: map[component.ID]extension.Extension{
				nopObsID:     nopObs,
				nopObsvbleID: nopObsvble,
			},
			watchObservers: []component.ID{nopObsID, nopObsvbleID},
			expectedError:  `extension "nop_observer" in watch_observers is not an observer`,
		},
		{name: "mixed non-observables ids with names",
			extensions: map[component.ID]extension.Extension{
				nopObsIDWithName:     nopObsWithName,
				nopObsvbleIDWithName: nopObsvbleWithName,
			},
			watchObservers: []component.ID{nopObsIDWithName, nopObsvbleIDWithName},
			expectedError:  `extension "nop_observer/with_name" in watch_observers is not an observer`,
		},
		{name: "only missing extension",
			extensions: map[component.ID]extension.Extension{
				nopObsvbleID: nopObsvble,
			},
			watchObservers: []component.ID{nopObsID},
			expectedError:  `failed to find observer "nop_observer" as a configured extension`,
		},
		{name: "happy path",
			extensions: map[component.ID]extension.Extension{
				nopObsvbleID:         nopObsvble,
				nopObsvbleIDWithName: nopObsvbleWithName,
			},
			watchObservers: []component.ID{nopObsvbleID, nopObsvbleIDWithName},
			expectedObservables: map[component.ID]observer.Observable{
				nopObsvbleID:         nopObsvble,
				nopObsvbleIDWithName: nopObsvbleWithName,
			},
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			rcs := otelcolreceiver.CreateSettings{
				TelemetrySettings: component.TelemetrySettings{
					Logger:         zap.NewNop(),
					TracerProvider: trace.NewNoopTracerProvider(),
					MeterProvider:  noop.NewMeterProvider(),
				},
			}
			host := mockHost{extensions: test.extensions}
			cfg := &Config{WatchObservers: test.watchObservers}
			receiver, err := newDiscoveryReceiver(rcs, cfg, consumertest.NewNop())
			require.NoError(t, err)
			require.NotNil(t, receiver)

			observables, err := receiver.observablesFromHost(host)
			if test.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
				require.Nil(t, observables)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedObservables, observables)
			}
		})
	}
}

type mockHost struct {
	component.Host
	extensions map[component.ID]extension.Extension
}

func (mh mockHost) GetFactory(_ component.Kind, _ component.Type) component.Factory {
	return nil
}

func (mh mockHost) GetExtensions() map[component.ID]extension.Extension {
	return mh.extensions
}

type nopObserver struct{}

var _ extension.Extension = (*nopObserver)(nil)
var _ observer.Observable = (*nopObservable)(nil)

func (m nopObserver) Start(_ context.Context, _ component.Host) error {
	return nil
}
func (m nopObserver) Shutdown(_ context.Context) error {
	return nil
}

type nopObservable struct {
	nopObserver
}

func (m nopObservable) ListAndWatch(_ observer.Notify) {}
func (m nopObservable) Unsubscribe(_ observer.Notify)  {}
