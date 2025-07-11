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
	"path"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

func TestEndpointToPLogsHappyPath(t *testing.T) {
	for _, tt := range []struct {
		endpoint      observer.Endpoint
		expectedPLogs plog.Logs
		name          string
	}{
		{
			name:     "pod",
			endpoint: podEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
				entityIDAttr.Map().PutStr("service.name", "my-mysql")
				entityIDAttr.Map().PutStr("k8s.pod.uid", "uid")
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "pod.endpoint.id")
				attrs.PutStr("endpoint", "pod.target")
				attrs.PutStr("k8s.pod.name", "my-mysql-0")
				attrs.PutStr("k8s.namespace.name", "namespace")
				attrs.PutStr("type", "pod")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
		{
			name:     "port",
			endpoint: portEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
				entityIDAttr.Map().PutStr("service.name", "redis-cart")
				entityIDAttr.Map().PutStr("k8s.pod.uid", "uid")
				entityIDAttr.Map().PutInt("source.port", 1)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "port.endpoint.id")
				attrs.PutStr("endpoint", "port.target")
				attrs.PutStr("k8s.pod.name", "redis-cart-657b69bb49-8csql")
				attrs.PutStr("k8s.namespace.name", "namespace")
				attrs.PutStr("type", "port")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
		{
			name:     "hostport",
			endpoint: hostportEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
				entityIDAttr.Map().PutInt("source.port", 1)
				entityIDAttr.Map().PutStr("service.name", "process.name")
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "hostport.endpoint.id")
				attrs.PutStr("endpoint", "hostport.target")
				attrs.PutStr("process.executable.name", "process.name")
				attrs.PutStr("type", "hostport")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
		{
			name:     "container",
			endpoint: containerEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
				entityIDAttr.Map().PutStr("service.name", "container.name")
				entityIDAttr.Map().PutStr("container.id", "container.id")
				entityIDAttr.Map().PutInt("source.port", 1)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "container.endpoint.id")
				attrs.PutStr("endpoint", "container.target")
				attrs.PutStr("container.name", "container.name")
				attrs.PutStr("type", "container")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
		{
			name:     "k8s.node",
			endpoint: k8sNodeEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
				entityIDAttr.Map().PutStr("k8s.node.uid", "uid")
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "k8s.node.endpoint.id")
				attrs.PutStr("endpoint", "k8s.node.target")
				attrs.PutStr("k8s.node.name", "k8s.node.name")
				attrs.PutStr("type", "k8s.node")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(test.endpoint.ID, map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityEvents(
				component.MustNewIDWithName("observer_type", "observer.name"),
				[]observer.Endpoint{test.endpoint}, corr, t0, experimentalmetricmetadata.EventTypeState,
			)
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, events.Len())

			require.NoError(t, plogtest.CompareLogs(test.expectedPLogs, events.ConvertAndMoveToLogs()))
		})
	}
}

func TestEndpointToPLogsInvalidEndpoints(t *testing.T) {
	for _, tt := range []struct {
		name          string
		endpoint      observer.Endpoint
		expectedPLogs plog.Logs
		expectedError string
	}{
		{
			name: "nil details",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: nil,
			},
			expectedError: `endpoint "endpoint.id" has no details`,
		},
		{
			name: "empty details env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: emptyDetailsEnv{},
			},
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "endpoint.id")
				attrs.PutStr("endpoint", "endpoint.target")
				attrs.PutStr("type", "empty.details.env")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
		{
			name: "unexpected pod field in env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: unexpectedPodInEnv{},
			},
			expectedError: `failed determining attributes for "endpoint.id": failed parsing port pod env "not a map"`,
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(test.endpoint.ID, map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityEvents(
				component.MustNewIDWithName("observer_type", "observer.name"),
				[]observer.Endpoint{test.endpoint}, corr, t0, experimentalmetricmetadata.EventTypeState,
			)
			if test.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, events.Len())
			require.NoError(t, plogtest.CompareLogs(test.expectedPLogs, events.ConvertAndMoveToLogs()))

			// Validate entity_delete event
			expectedDeleteEvent := test.expectedPLogs
			lr := expectedDeleteEvent.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.Attributes().PutStr(discovery.OtelEntityEventTypeAttr, discovery.OtelEntityEventTypeDelete)
			lr.Attributes().Remove(discovery.OtelEntityAttributesAttr)
			events, failed, err = entityEvents(
				component.MustNewIDWithName("observer_type", "observer.name"),
				[]observer.Endpoint{test.endpoint}, corr, t0, experimentalmetricmetadata.EventTypeDelete,
			)
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, events.Len())
			require.NoError(t, plogtest.CompareLogs(expectedDeleteEvent, events.ConvertAndMoveToLogs()))
		})
	}
}

func FuzzEndpointToPlogs(f *testing.F) {
	f.Add("observer_type", "observer.name",
		"port.endpoint.id", "port.target", "port.name", "pod.name", "uid", "label.value",
		"annotation.one", "annotation.value.one", "annotation.two", "annotation.value.two",
		"namespace", "transport", uint16(1))
	f.Add(discovery.NoType.Type().String(), "", "", "", "", "", "", "", "", "", "", "", "", "", uint16(0))
	f.Fuzz(func(t *testing.T, observerType, observerName,
		endpointID, target, portName, podName, uid, labelValue,
		annotationOne, annotationValueOne, annotationTwo, annotationValueTwo,
		namespace, transport string, port uint16) {
		require.NotPanics(t, func() {
			observerTypeSanitized, err := component.NewType(observerType)
			if err != nil {
				observerTypeSanitized = discovery.NoType.Type()
			}
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(observer.EndpointID(endpointID), map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityEvents(
				component.MustNewIDWithName(observerTypeSanitized.String(), observerName), []observer.Endpoint{
					{
						ID:     observer.EndpointID(endpointID),
						Target: target,
						Details: &observer.Port{
							Name: portName,
							Pod: observer.Pod{
								Name: podName,
								UID:  uid,
								Labels: map[string]string{
									"app": labelValue,
								},
								Annotations: map[string]string{
									annotationOne: annotationValueOne,
									annotationTwo: annotationValueTwo,
								},
								Namespace: namespace,
							},
							Port:      port,
							Transport: observer.Transport(transport),
						},
					},
				}, corr, t0, experimentalmetricmetadata.EventTypeState,
			)

			expectedLogs := expectedPLogs()
			lr := expectedLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.SetTimestamp(pcommon.NewTimestampFromTime(t0))
			entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
			entityIDAttr.Map().PutStr("service.name", labelValue)
			entityIDAttr.Map().PutStr("k8s.pod.uid", uid)
			entityIDAttr.Map().PutInt("source.port", int64(port))
			attrs := lr.Attributes().PutEmptyMap(discovery.OtelEntityAttributesAttr)
			attrs.PutStr(discovery.EndpointIDAttr, endpointID)
			attrs.PutStr(discovery.StatusAttr, "successful")
			attrs.PutStr(observerNameAttr, observerName)
			attrs.PutStr(observerTypeAttr, observerTypeSanitized.String())
			attrs.PutStr("endpoint", target)

			attrs.PutStr("k8s.pod.name", podName)
			attrs.PutStr("k8s.namespace.name", namespace)
			attrs.PutStr("type", "port")
			require.Equal(t, 1, events.Len())

			require.NoError(t, plogtest.CompareLogs(expectedLogs, events.ConvertAndMoveToLogs()))
			require.NoError(t, err)
			require.Zero(t, failed)
		})
	})
}

var (
	t0 = time.Unix(0, 0).UTC()

	podEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("pod.endpoint.id"),
		Target: "pod.target",
		Details: &observer.Pod{
			Name: "my-mysql-0",
			UID:  "uid",
			Labels: map[string]string{
				"label.one": "value.one",
				"label.two": "value.two",
			},
			Annotations: map[string]string{
				"annotation.one": "value.one",
				"annotation.two": "value.two",
			},
			Namespace: "namespace",
		},
	}

	portEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("port.endpoint.id"),
		Target: "port.target",
		Details: &observer.Port{
			Name: "port.name",
			Pod: observer.Pod{
				Name: "redis-cart-657b69bb49-8csql",
				UID:  "uid",
				Labels: map[string]string{
					"label.one": "value.one",
					"label.two": "value.two",
				},
				Annotations: map[string]string{
					"annotation.one": "value.one",
					"annotation.two": "value.two",
				},
				Namespace: "namespace",
			},
			Port:      1,
			Transport: "transport",
		},
	}

	hostportEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("hostport.endpoint.id"),
		Target: "hostport.target",
		Details: &observer.HostPort{
			ProcessName: "process.name",
			Command:     "command",
			Port:        1,
			Transport:   "transport",
			IsIPv6:      true,
		},
	}

	containerEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("container.endpoint.id"),
		Target: "container.target",
		Details: &observer.Container{
			Name:          "container.name",
			Image:         "image",
			Tag:           "tag",
			Port:          1,
			AlternatePort: 2,
			Command:       "command",
			ContainerID:   "container.id",
			Host:          "host",
			Transport:     "transport",
			Labels: map[string]string{
				"label.one": "value.one",
				"label.two": "value.two",
			},
		},
	}

	k8sNodeEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("k8s.node.endpoint.id"),
		Target: "k8s.node.target",
		Details: &observer.K8sNode{
			Name:        "k8s.node.name",
			UID:         "uid",
			Hostname:    "host.name",
			ExternalIP:  "external.ip",
			InternalIP:  "internal.ip",
			ExternalDNS: "external.dns",
			InternalDNS: "internal.dns",
			Annotations: map[string]string{
				"annotation.one": "value.one",
				"annotation.two": "value.two",
			},
			Labels: map[string]string{
				"label.one": "value.one",
				"label.two": "value.two",
			},
			KubeletEndpointPort: 1,
		},
	}
)

func expectedPLogs() plog.Logs {
	plogs := plog.NewLogs()
	scopeLog := plogs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	scopeLog.Scope().Attributes().PutBool(discovery.OtelEntityEventAsLogAttr, true)
	lr := scopeLog.LogRecords().AppendEmpty()
	lr.Attributes().PutStr(discovery.OtelEntityTypeAttr, entityType)
	lr.Attributes().PutStr(discovery.OtelEntityEventTypeAttr, discovery.OtelEntityEventTypeState)
	id := lr.Attributes().PutEmptyMap(discovery.OtelEntityIDAttr)
	id.PutStr("service.name", "unknown")
	attrs := lr.Attributes().PutEmptyMap(discovery.OtelEntityAttributesAttr)
	attrs.PutStr(observerNameAttr, "observer.name")
	attrs.PutStr(observerTypeAttr, "observer_type")
	return plogs
}

var (
	_ observer.EndpointDetails = (*emptyDetailsEnv)(nil)
	_ observer.EndpointDetails = (*unexpectedPodInEnv)(nil)
)

type emptyDetailsEnv struct{}

func (n emptyDetailsEnv) Env() observer.EndpointEnv {
	return nil
}

func (n emptyDetailsEnv) Type() observer.EndpointType {
	return "empty.details.env"
}

type unexpectedPodInEnv struct{}

func (n unexpectedPodInEnv) Env() observer.EndpointEnv {
	return map[string]any{
		"pod": "not a map",
	}
}

func (n unexpectedPodInEnv) Type() observer.EndpointType {
	return observer.PortType
}

func TestUpdateEndpoints(t *testing.T) {
	tests := []struct {
		name                 string
		config               string
		endpoints            []observer.Endpoint
		expectedCorrelations int
	}{
		{
			name:                 "no_endpoints",
			config:               "config.yaml",
			endpoints:            nil,
			expectedCorrelations: 0,
		},
		{
			name:                 "no_matching_endpoints",
			config:               "config.yaml",
			endpoints:            []observer.Endpoint{containerEndpoint},
			expectedCorrelations: 0,
		},
		{
			name:   "one_matching_endpoint",
			config: "config.yaml",
			endpoints: []observer.Endpoint{
				{
					ID:     observer.EndpointID("container.endpoint.id"),
					Target: "container.target",
					Details: &observer.Container{
						Name:        "Redis-app",
						Image:       "image",
						Tag:         "tag",
						Port:        1,
						Command:     "command",
						ContainerID: "container.id",
						Host:        "host",
					},
				},
			},
			expectedCorrelations: 1,
		},
		{
			name:   "multiple_matching_endpoints",
			config: "conflicting_rules.yaml",
			endpoints: []observer.Endpoint{
				{
					ID:     observer.EndpointID("container.endpoint.id"),
					Target: "container.target",
					Details: &observer.Container{
						Name:  "app",
						Image: "image1",
						Port:  2,
					},
				},
				{
					ID:     observer.EndpointID("container.endpoint.id"),
					Target: "container.target",
					Details: &observer.Container{
						Name:  "app",
						Image: "image2",
						Port:  1,
					},
				},
			},
			expectedCorrelations: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := confmaptest.LoadConf(path.Join(".", "testdata", tt.config))
			require.NoError(t, err)
			cm, err := config.Sub(typeStr)
			require.NoError(t, err)
			cfg := createDefaultConfig().(*Config)
			require.NoError(t, cm.Unmarshal(&cfg))

			logger := zap.NewNop()
			et := newEndpointTracker(nil, cfg, logger, nil, newCorrelationStore(logger, cfg.CorrelationTTL))
			et.updateEndpoints(tt.endpoints, component.MustNewIDWithName("observer_type", "observer.name"))

			var correlationsCount int
			et.correlations.correlations.Range(func(_, _ interface{}) bool {
				correlationsCount++
				return true
			})
			require.Equal(t, tt.expectedCorrelations, correlationsCount)
		})
	}
}

func TestEntityEmittingLifecycle(t *testing.T) {
	logger := zap.NewNop()
	cfg := createDefaultConfig().(*Config)
	cfg.Receivers = map[component.ID]ReceiverEntry{
		component.MustNewIDWithName("fake_receiver", ""): {
			Rule: mustNewRule(`type == "port" && pod.name matches "(?i)redis" && port == 1`),
		},
	}

	ch := make(chan plog.Logs, 10)
	obsID := component.MustNewIDWithName("fake_observer", "")
	obs := &fakeObservable{}
	corr := newCorrelationStore(logger, cfg.CorrelationTTL)
	et := &endpointTracker{
		config:       cfg,
		observables:  map[component.ID]observer.Observable{obsID: obs},
		logger:       logger,
		pLogs:        ch,
		correlations: corr,
		emitInterval: 20 * time.Millisecond,
		stopCh:       make(chan struct{}),
	}
	et.start()
	defer et.stop()

	// Wait for obs.ListAndWatch called asynchronously in the et.start()
	require.Eventually(t, func() bool {
		obs.lock.Lock()
		defer obs.lock.Unlock()
		return obs.onAdd != nil
	}, 200*time.Millisecond, 10*time.Millisecond)

	obs.onAdd([]observer.Endpoint{portEndpoint})

	// Ensure that no entities are emitted without discovery.status attribute.
	time.Sleep(30 * time.Millisecond)
	require.Empty(t, ch)

	// Once the status attribute is set, the entity should be emitted.
	corr.attrs.Store(portEndpoint.ID, map[string]string{discovery.StatusAttr: "successful"})

	// Wait for at least 2 entity events to be emitted to confirm periodic emitting is working.
	require.Eventually(t, func() bool { return len(ch) >= 2 }, 1*time.Second, 60*time.Millisecond)

	gotLogs := <-ch
	require.Equal(t, 1, gotLogs.LogRecordCount())
	expectedEvents, failed, err := entityEvents(obsID, []observer.Endpoint{portEndpoint}, corr, t0, experimentalmetricmetadata.EventTypeState)
	require.NoError(t, err)
	require.Zero(t, failed)
	require.NoError(t, plogtest.CompareLogs(expectedEvents.ConvertAndMoveToLogs(), gotLogs, plogtest.IgnoreTimestamp()))

	// Remove the endpoint.
	obs.onRemove([]observer.Endpoint{portEndpoint})

	// Wait for an entity delete event.
	expectedEvents, failed, err = entityEvents(
		component.MustNewIDWithName("observer_type", "observer.name"),
		[]observer.Endpoint{portEndpoint}, corr, t0, experimentalmetricmetadata.EventTypeDelete,
	)
	require.NoError(t, err)
	require.Zero(t, failed)
	expectedLogs := expectedEvents.ConvertAndMoveToLogs()
	waitChan := make(chan struct{})
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		logs := <-ch
		if assert.NoError(c, plogtest.CompareLogs(expectedLogs, logs, plogtest.IgnoreTimestamp())) {
			close(waitChan)
		}
	}, 1*time.Second, 50*time.Millisecond)

	// Ensure that entities are not emitted anymore
	<-waitChan
	assert.Empty(t, ch)
}

type fakeObservable struct {
	onAdd    func([]observer.Endpoint)
	onRemove func([]observer.Endpoint)
	lock     sync.Mutex
}

func (f *fakeObservable) ListAndWatch(notify observer.Notify) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.onAdd = notify.OnAdd
	f.onRemove = notify.OnRemove
}

func (f *fakeObservable) Unsubscribe(observer.Notify) {}

func TestEntityStateEvents(t *testing.T) {
	logger := zap.NewNop()
	cfg := createDefaultConfig().(*Config)
	cfg.Receivers = map[component.ID]ReceiverEntry{
		component.MustNewIDWithName("redis", ""): {
			Rule: mustNewRule(`type == "port" && pod.name matches "(?i)redis" && port == 1`),
		},
	}

	cStore := newCorrelationStore(logger, cfg.CorrelationTTL)
	cStore.UpdateAttrs(portEndpoint.ID, map[string]string{
		discovery.ReceiverTypeAttr: "redis",
		"service.type":             "redis",
		discovery.StatusAttr:       "successful",
		"attr1":                    "val1",
		"attr2":                    "val2",
	})

	events, failed, err := entityEvents(component.MustNewIDWithName("observer_type", "observer.name"),
		[]observer.Endpoint{portEndpoint}, cStore, t0, experimentalmetricmetadata.EventTypeState)
	require.NoError(t, err)
	require.Zero(t, failed)
	require.Equal(t, 1, events.Len())

	event := events.At(0)
	assert.Equal(t, experimentalmetricmetadata.EventTypeState, event.EventType())
	assert.Equal(t, t0, event.Timestamp().AsTime())
	assert.Equal(t, map[string]any{
		"service.type": "redis",
		"service.name": "redis-cart",
		"k8s.pod.uid":  "uid",
		"source.port":  int64(1),
	}, event.ID().AsRaw())
	assert.Equal(t, map[string]any{
		observerNameAttr:          "observer.name",
		observerTypeAttr:          "observer_type",
		discovery.StatusAttr:      "successful",
		"endpoint":                "port.target",
		"discovery.receiver.type": "redis",
		"discovery.endpoint.id":   "port.endpoint.id",
		"k8s.pod.name":            "redis-cart-657b69bb49-8csql",
		"k8s.namespace.name":      "namespace",
		"type":                    "port",
		"attr1":                   "val1",
		"attr2":                   "val2",
	}, event.EntityStateDetails().Attributes().AsRaw())
}

func TestEntityDeleteEvents(t *testing.T) {
	cStore := newCorrelationStore(zap.NewNop(), time.Hour)
	cStore.attrs.Store(portEndpoint.ID, map[string]string{discovery.StatusAttr: "successful"})
	events, failed, err := entityEvents(component.MustNewIDWithName("observer_type", "observer.name"),
		[]observer.Endpoint{portEndpoint}, cStore, t0, experimentalmetricmetadata.EventTypeDelete)
	require.Zero(t, failed)
	require.NoError(t, err)
	require.Equal(t, 1, events.Len())

	event := events.At(0)
	assert.Equal(t, experimentalmetricmetadata.EventTypeDelete, event.EventType())
	assert.Equal(t, t0, event.Timestamp().AsTime())
	assert.Equal(t, map[string]any{
		"service.name": "redis-cart",
		sourcePortAttr: int64(1),
		"k8s.pod.uid":  "uid",
	}, event.ID().AsRaw())
}

func TestDeduceServiceName(t *testing.T) {
	tests := []struct {
		name         string
		endpointType observer.EndpointType
		endpointEnv  observer.EndpointEnv
		expected     string
	}{
		{
			name:         "k8s.pod.name",
			endpointType: observer.PodType,
			endpointEnv: observer.EndpointEnv{
				"pod": observer.EndpointEnv{
					"name": "my-daemonset-f6pxf",
				},
				"name": "my-name",
			},
			expected: "my-daemonset",
		},
		{
			name:         "labels.app",
			endpointType: observer.PodType,
			endpointEnv: observer.EndpointEnv{
				"labels": map[string]string{
					"app": "my-app",
				},
			},
			expected: "my-app",
		},
		{
			name:         "pod-port-new-k8s-labels",
			endpointType: observer.PortType,
			endpointEnv: observer.EndpointEnv{
				"pod": observer.EndpointEnv{
					"labels": map[string]string{
						"app.kubernetes.io/name": "my-app-new-name",
						"app":                    "my-app-old-name",
					},
				},
				"process_name": "my-process",
			},
			expected: "my-app-new-name",
		},
		{
			name:         "pod-port-old-k8s-labels",
			endpointType: observer.PortType,
			endpointEnv: observer.EndpointEnv{
				"pod": observer.EndpointEnv{
					"labels": map[string]string{
						"app": "my-app-old-name",
					},
				},
				"process_name": "my-process",
			},
			expected: "my-app-old-name",
		},
		{
			name:         "name",
			endpointType: observer.ContainerType,
			endpointEnv: observer.EndpointEnv{
				"name": "my-name",
			},
			expected: "my-name",
		},
		{
			name:         "process_name",
			endpointType: observer.HostPortType,
			endpointEnv: observer.EndpointEnv{
				"process_name": "my-process",
			},
			expected: "my-process",
		},
		{
			name:         "empty",
			endpointType: observer.ContainerType,
			endpointEnv:  observer.EndpointEnv{},
			expected:     "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractServiceName(tt.endpointType, tt.endpointEnv))
		})
	}
}
