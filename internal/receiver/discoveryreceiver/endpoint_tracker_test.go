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
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
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
				plogs := expectedPLogs("pod.endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutStr("annotation.one", "value.one")
				annotationsMap.PutStr("annotation.two", "value.two")
				attrs.PutStr("endpoint", "pod.target")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("name", "pod.name")
				attrs.PutStr("namespace", "namespace")
				attrs.PutStr("type", "pod")
				attrs.PutStr("uid", "uid")
				return plogs
			}(),
		},
		{
			name:     "port",
			endpoint: portEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("port.endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr("endpoint", "port.target")
				attrs.PutStr("name", "port.name")

				podEnvMap := attrs.PutEmptyMap("pod")
				annotationsMap := podEnvMap.PutEmptyMap("annotations")
				annotationsMap.PutStr("annotation.one", "value.one")
				annotationsMap.PutStr("annotation.two", "value.two")
				labelsMap := podEnvMap.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				podEnvMap.PutStr("name", "pod.name")
				podEnvMap.PutStr("namespace", "namespace")
				podEnvMap.PutStr("uid", "uid")

				attrs.PutInt("port", 1)
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "port")
				return plogs
			}(),
		},
		{
			name:     "hostport",
			endpoint: hostportEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("hostport.endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr("command", "command")
				attrs.PutStr("endpoint", "hostport.target")
				attrs.PutBool("is_ipv6", true)
				attrs.PutInt("port", 1)
				attrs.PutStr("process_name", "process.name")
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "hostport")
				return plogs
			}(),
		},
		{
			name:     "container",
			endpoint: containerEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("container.endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutInt("alternate_port", 2)
				attrs.PutStr("command", "command")
				attrs.PutStr("container_id", "container.id")
				attrs.PutStr("endpoint", "container.target")
				attrs.PutStr("host", "host")
				attrs.PutStr("image", "image")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("name", "container.name")
				attrs.PutInt("port", 1)
				attrs.PutStr("tag", "tag")
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "container")
				return plogs
			}(),
		},
		{
			name:     "k8s.node",
			endpoint: k8sNodeEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("k8s.node.endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutStr("annotation.one", "value.one")
				annotationsMap.PutStr("annotation.two", "value.two")
				attrs.PutStr("endpoint", "k8s.node.target")
				attrs.PutStr("external_dns", "external.dns")
				attrs.PutStr("external_ip", "external.ip")
				attrs.PutStr("hostname", "host.name")
				attrs.PutStr("internal_dns", "internal.dns")
				attrs.PutStr("internal_ip", "internal.ip")
				attrs.PutInt("kubelet_endpoint_port", 1)
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("name", "k8s.node.name")
				attrs.PutStr("type", "k8s.node")
				attrs.PutStr("uid", "uid")
				return plogs
			}(),
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			plogs, failed, err := endpointToPLogs(
				component.MustNewIDWithName("observer_type", "observer.name"),
				"event.type", []observer.Endpoint{test.endpoint}, t0,
			)
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, plogs.LogRecordCount())

			require.NoError(t, plogtest.CompareLogs(test.expectedPLogs, plogs))
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
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr("endpoint", "endpoint.target")
				return plogs
			}(),
		},
		{
			name: "empty details env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: emptyDetailsEnv{},
			},
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr("endpoint", "endpoint.target")
				attrs.PutStr("type", "empty.details.env")
				return plogs
			}(),
		},
		{
			name: "unexpected labels and annotations in env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: unexpectedLabelsAndAnnotations{t: observer.EndpointType("unexpected.env")},
			},
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs("endpoint.id")
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutBool("annotations", false)
				attrs.PutStr("endpoint", "endpoint.target")
				attrs.PutBool("labels", true)
				attrs.PutStr("type", "unexpected.env")
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
		{
			name: "unexpected pod labels and annotations in env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: unexpectedLabelsAndAnnotations{t: observer.PodType},
			},
			expectedError: `failed determining attributes for "endpoint.id": failed parsing pod env attributes`,
		},
		{
			name: "unexpected k8s.node labels and annotations in env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: unexpectedLabelsAndAnnotations{t: observer.K8sNodeType},
			},
			expectedError: `failed determining attributes for "endpoint.id": failed parsing k8s.node env attributes`,
		},
		{
			name: "unexpected k8s.node labels and annotations in env",
			endpoint: observer.Endpoint{
				ID:      "endpoint.id",
				Target:  "endpoint.target",
				Details: unexpectedLabelsAndAnnotations{t: observer.ContainerType},
			},
			expectedError: `failed determining attributes for "endpoint.id": failed parsing container env attributes`,
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			// Validate entity_state event
			plogs, failed, err := endpointToPLogs(
				component.MustNewIDWithName("observer_type", "observer.name"),
				addedState, []observer.Endpoint{test.endpoint}, t0,
			)
			if test.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, plogs.LogRecordCount())
			require.NoError(t, plogtest.CompareLogs(test.expectedPLogs, plogs))

			// Validate entity_delete event
			expectedDeleteEvent := test.expectedPLogs
			lr := expectedDeleteEvent.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.Attributes().PutStr(discovery.OtelEntityEventTypeAttr, discovery.OtelEntityEventTypeDelete)
			lr.Attributes().Remove(discovery.OtelEntityAttributesAttr)
			plogs, failed, err = endpointToPLogs(
				component.MustNewIDWithName("observer_type", "observer.name"),
				removedState, []observer.Endpoint{test.endpoint}, t0,
			)
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, plogs.LogRecordCount())
			require.NoError(t, plogtest.CompareLogs(expectedDeleteEvent, plogs))
		})
	}
}

func FuzzEndpointToPlogs(f *testing.F) {
	f.Add("observer_type", "observer.name",
		"port.endpoint.id", "port.target", "port.name", "pod.name", "uid",
		"label.one", "label.value.one", "label.two", "label.value.two",
		"annotation.one", "annotation.value.one", "annotation.two", "annotation.value.two",
		"namespace", "transport", uint16(1))
	f.Add(discovery.NoType.Type().String(), "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", uint16(0))
	f.Fuzz(func(t *testing.T, observerType, observerName,
		endpointID, target, portName, podName, uid,
		labelOne, labelValueOne, labelTwo, labelValueTwo,
		annotationOne, annotationValueOne, annotationTwo, annotationValueTwo,
		namespace, transport string, port uint16) {
		require.NotPanics(t, func() {
			observerTypeSanitized, err := component.NewType(observerType)
			if err != nil {
				observerTypeSanitized = discovery.NoType.Type()
			}
			plogs, failed, err := endpointToPLogs(
				component.MustNewIDWithName(observerTypeSanitized.String(), observerName), addedState, []observer.Endpoint{
					{
						ID:     observer.EndpointID(endpointID),
						Target: target,
						Details: &observer.Port{
							Name: portName,
							Pod: observer.Pod{
								Name: podName,
								UID:  uid,
								Labels: map[string]string{
									labelOne: labelValueOne,
									labelTwo: labelValueTwo,
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
				}, t0,
			)

			expectedLogs := expectedPLogs(endpointID)
			lr := expectedLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.SetTimestamp(pcommon.NewTimestampFromTime(t0))
			attrs := lr.Attributes().PutEmptyMap(discovery.OtelEntityAttributesAttr)
			attrs.PutStr(observerNameAttr, observerName)
			attrs.PutStr(observerTypeAttr, observerTypeSanitized.String())
			attrs.PutStr("endpoint", target)
			attrs.PutStr("name", portName)

			podEnvMap := attrs.PutEmptyMap("pod")
			annotationsMap := podEnvMap.PutEmptyMap("annotations")
			annotationsMap.PutStr(annotationOne, annotationValueOne)
			annotationsMap.PutStr(annotationTwo, annotationValueTwo)
			labelsMap := podEnvMap.PutEmptyMap("labels")
			labelsMap.PutStr(labelOne, labelValueOne)
			labelsMap.PutStr(labelTwo, labelValueTwo)
			podEnvMap.PutStr("name", podName)
			podEnvMap.PutStr("namespace", namespace)
			podEnvMap.PutStr("uid", uid)
			attrs.PutInt("port", int64(port))
			attrs.PutStr("transport", transport)
			attrs.PutStr("type", "port")
			require.Equal(t, 1, plogs.LogRecordCount())

			require.NoError(t, plogtest.CompareLogs(expectedLogs, plogs))
			require.NoError(t, err)
			require.Zero(t, failed)
		})
	})
}

var (
	t0 = time.Unix(0, 0)

	podEndpoint = observer.Endpoint{
		ID:     observer.EndpointID("pod.endpoint.id"),
		Target: "pod.target",
		Details: &observer.Pod{
			Name: "pod.name",
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
				Name: "pod.name",
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

func expectedPLogs(endpointID string) plog.Logs {
	plogs := plog.NewLogs()
	scopeLog := plogs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	scopeLog.Scope().Attributes().PutBool(discovery.OtelEntityEventAsLogAttr, true)
	lr := scopeLog.LogRecords().AppendEmpty()
	lr.Attributes().PutStr(discovery.OtelEntityEventTypeAttr, discovery.OtelEntityEventTypeState)
	lr.Attributes().PutEmptyMap(discovery.OtelEntityIDAttr).PutStr(discovery.EndpointIDAttr, endpointID)
	attrs := lr.Attributes().PutEmptyMap(discovery.OtelEntityAttributesAttr)
	attrs.PutStr(observerNameAttr, "observer.name")
	attrs.PutStr(observerTypeAttr, "observer_type")
	return plogs
}

var (
	_ observer.EndpointDetails = (*emptyDetailsEnv)(nil)
	_ observer.EndpointDetails = (*unexpectedLabelsAndAnnotations)(nil)
	_ observer.EndpointDetails = (*unexpectedPodInEnv)(nil)
)

type emptyDetailsEnv struct{}

func (n emptyDetailsEnv) Env() observer.EndpointEnv {
	return nil
}

func (n emptyDetailsEnv) Type() observer.EndpointType {
	return "empty.details.env"
}

type unexpectedLabelsAndAnnotations struct {
	t observer.EndpointType
}

func (n unexpectedLabelsAndAnnotations) Env() observer.EndpointEnv {
	return map[string]any{
		"labels":      true,
		"annotations": false,
	}
}

func (n unexpectedLabelsAndAnnotations) Type() observer.EndpointType {
	return n.t
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
			require.NoError(t, component.UnmarshalConfig(cm, cfg))

			logger := zap.NewNop()
			et := newEndpointTracker(nil, cfg, logger, make(chan plog.Logs), newCorrelationStore(logger, cfg.CorrelationTTL))
			et.updateEndpoints(tt.endpoints, addedState, component.MustNewIDWithName("observer_type", "observer.name"))

			var correlationsCount int
			et.correlations.(*store).correlations.Range(func(_, _ interface{}) bool {
				correlationsCount++
				return true
			})
			require.Equal(t, tt.expectedCorrelations, correlationsCount)
		})
	}
}

func TestPeriodicEntityEmitting(t *testing.T) {
	logger := zap.NewNop()
	cfg := createDefaultConfig().(*Config)
	cfg.LogEndpoints = true
	cfg.Receivers = map[component.ID]ReceiverEntry{
		component.MustNewIDWithName("fake_receiver", ""): {
			Rule: mustNewRule(`type == "port" && pod.name == "pod.name" && port == 1`),
		},
	}

	ch := make(chan plog.Logs, 10)
	obsID := component.MustNewIDWithName("fake_observer", "")
	obs := &fakeObservable{}
	et := &endpointTracker{
		config:       cfg,
		observables:  map[component.ID]observer.Observable{obsID: obs},
		logger:       logger,
		pLogs:        ch,
		correlations: newCorrelationStore(logger, cfg.CorrelationTTL),
		emitInterval: 100 * time.Millisecond,
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

	// Wait for at least 2 entity events to be emitted
	require.Eventually(t, func() bool { return len(ch) >= 2 }, 1*time.Second, 50*time.Millisecond)

	gotLogs := <-ch
	require.Equal(t, 1, gotLogs.LogRecordCount())
	// TODO: Use plogtest.IgnoreTimestamp once available
	expectedLogs, failed, err := endpointToPLogs(obsID, addedState, []observer.Endpoint{portEndpoint},
		gotLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Timestamp().AsTime())
	require.NoError(t, err)
	require.Zero(t, failed)
	require.NoError(t, plogtest.CompareLogs(expectedLogs, gotLogs))
}

type fakeObservable struct {
	onAdd func([]observer.Endpoint)
	lock  sync.Mutex
}

func (f *fakeObservable) ListAndWatch(notify observer.Notify) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.onAdd = notify.OnAdd
}

func (f *fakeObservable) Unsubscribe(observer.Notify) {}
