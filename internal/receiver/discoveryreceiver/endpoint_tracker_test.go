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
				entityIDAttr.Map().PutStr("k8s.pod.uid", "uid")
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "pod.endpoint.id")
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutStr("annotation.one", "value.one")
				annotationsMap.PutStr("annotation.two", "value.two")
				attrs.PutStr("endpoint", "pod.target")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("k8s.pod.name", "my-mysql-0")
				attrs.PutStr("k8s.namespace.name", "namespace")
				attrs.PutStr("type", "pod")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "my-mysql")
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
				entityIDAttr.Map().PutStr("k8s.pod.uid", "uid")
				entityIDAttr.Map().PutInt("source.port", 1)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "port.endpoint.id")
				attrs.PutStr("endpoint", "port.target")
				attrs.PutStr("name", "port.name")
				attrs.PutStr("k8s.pod.name", "redis-cart-657b69bb49-8csql")
				attrs.PutStr("k8s.namespace.name", "namespace")
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutStr("annotation.one", "value.one")
				annotationsMap.PutStr("annotation.two", "value.two")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "port")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "redis-cart")
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
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "hostport.endpoint.id")
				attrs.PutStr("command", "command")
				attrs.PutStr("endpoint", "hostport.target")
				attrs.PutBool("is_ipv6", true)
				attrs.PutStr("process_name", "process.name")
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "hostport")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "process.name")
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
				entityIDAttr.Map().PutStr("container.id", "container.id")
				entityIDAttr.Map().PutInt("source.port", 1)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "container.endpoint.id")
				attrs.PutInt("alternate_port", 2)
				attrs.PutStr("command", "command")
				attrs.PutStr("endpoint", "container.target")
				attrs.PutStr("host", "host")
				attrs.PutStr("image", "image")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutStr("label.one", "value.one")
				labelsMap.PutStr("label.two", "value.two")
				attrs.PutStr("name", "container.name")
				attrs.PutStr("tag", "tag")
				attrs.PutStr("transport", "transport")
				attrs.PutStr("type", "container")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "container.name")
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
				attrs.PutStr("k8s.node.name", "k8s.node.name")
				attrs.PutStr("type", "k8s.node")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "unknown")
				attrs.PutStr(discovery.StatusAttr, "successful")
				return plogs
			}(),
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(test.endpoint.ID, map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityStateEvents(
				component.MustNewIDWithName("observer_type", "observer.name"),
				[]observer.Endpoint{test.endpoint}, corr, t0,
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
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "unknown")
				attrs.PutStr(discovery.StatusAttr, "successful")
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
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				entityAttrsAttr, _ := lr.Attributes().Get(discovery.OtelEntityAttributesAttr)
				attrs := entityAttrsAttr.Map()
				attrs.PutStr(discovery.EndpointIDAttr, "endpoint.id")
				attrs.PutBool("annotations", false)
				attrs.PutStr("endpoint", "endpoint.target")
				attrs.PutBool("labels", true)
				attrs.PutStr("type", "unexpected.env")
				attrs.PutStr("service.type", "unknown")
				attrs.PutStr("service.name", "unknown")
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
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(test.endpoint.ID, map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityStateEvents(
				component.MustNewIDWithName("observer_type", "observer.name"),
				[]observer.Endpoint{test.endpoint}, corr, t0,
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
			lr.Attributes().Remove(discovery.OtelEntityTypeAttr)
			lr.Attributes().Remove(discovery.OtelEntityAttributesAttr)
			events, failed, err = entityDeleteEvents([]observer.Endpoint{test.endpoint}, t0)
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, events.Len())
			require.NoError(t, plogtest.CompareLogs(expectedDeleteEvent, events.ConvertAndMoveToLogs()))
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
			corr := newCorrelationStore(zap.NewNop(), time.Hour)
			corr.attrs.Store(observer.EndpointID(endpointID), map[string]string{discovery.StatusAttr: "successful"})
			events, failed, err := entityStateEvents(
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
				}, corr, t0,
			)

			expectedLogs := expectedPLogs()
			lr := expectedLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.SetTimestamp(pcommon.NewTimestampFromTime(t0))
			entityIDAttr, _ := lr.Attributes().Get(discovery.OtelEntityIDAttr)
			entityIDAttr.Map().PutStr("k8s.pod.uid", uid)
			entityIDAttr.Map().PutInt("source.port", int64(port))
			attrs := lr.Attributes().PutEmptyMap(discovery.OtelEntityAttributesAttr)
			attrs.PutStr(discovery.EndpointIDAttr, endpointID)
			attrs.PutStr(discovery.StatusAttr, "successful")
			attrs.PutStr(observerNameAttr, observerName)
			attrs.PutStr(observerTypeAttr, observerTypeSanitized.String())
			attrs.PutStr("endpoint", target)
			attrs.PutStr("name", portName)

			annotationsMap := attrs.PutEmptyMap("annotations")
			annotationsMap.PutStr(annotationOne, annotationValueOne)
			annotationsMap.PutStr(annotationTwo, annotationValueTwo)
			labelsMap := attrs.PutEmptyMap("labels")
			labelsMap.PutStr(labelOne, labelValueOne)
			labelsMap.PutStr(labelTwo, labelValueTwo)
			attrs.PutStr("k8s.pod.name", podName)
			attrs.PutStr("k8s.namespace.name", namespace)
			attrs.PutStr("transport", transport)
			attrs.PutStr("type", "port")
			attrs.PutStr("service.type", "unknown")
			attrs.PutStr("service.name", portName)
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
	lr.Attributes().PutEmptyMap(discovery.OtelEntityIDAttr)
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
	expectedEvents, failed, err := entityStateEvents(obsID, []observer.Endpoint{portEndpoint}, corr, t0)
	require.NoError(t, err)
	require.Zero(t, failed)
	require.NoError(t, plogtest.CompareLogs(expectedEvents.ConvertAndMoveToLogs(), gotLogs, plogtest.IgnoreTimestamp()))

	// Remove the endpoint.
	obs.onRemove([]observer.Endpoint{portEndpoint})

	// Wait for an entity delete event.
	expectedEvents, failed, err = entityDeleteEvents([]observer.Endpoint{portEndpoint}, t0)
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
		discovery.StatusAttr:       "successful",
		"attr1":                    "val1",
		"attr2":                    "val2",
	})

	events, failed, err := entityStateEvents(component.MustNewIDWithName("observer_type", "observer.name"),
		[]observer.Endpoint{portEndpoint}, cStore, t0)
	require.NoError(t, err)
	require.Zero(t, failed)
	require.Equal(t, 1, events.Len())

	event := events.At(0)
	assert.Equal(t, experimentalmetricmetadata.EventTypeState, event.EventType())
	assert.Equal(t, t0, event.Timestamp().AsTime())
	assert.Equal(t, map[string]any{"k8s.pod.uid": "uid", "source.port": int64(1)}, event.ID().AsRaw())
	assert.Equal(t, map[string]any{
		observerNameAttr:     "observer.name",
		observerTypeAttr:     "observer_type",
		discovery.StatusAttr: "successful",
		"endpoint":           "port.target",
		"name":               "port.name",
		"annotations": map[string]any{
			"annotation.one": "value.one",
			"annotation.two": "value.two",
		},
		"labels": map[string]any{
			"label.one": "value.one",
			"label.two": "value.two",
		},
		"discovery.receiver.type": "redis",
		"discovery.endpoint.id":   "port.endpoint.id",
		"k8s.pod.name":            "redis-cart-657b69bb49-8csql",
		"k8s.namespace.name":      "namespace",
		"transport":               "transport",
		"type":                    "port",
		"attr1":                   "val1",
		"attr2":                   "val2",
		"service.type":            "redis",
		"service.name":            "redis-cart",
	}, event.EntityStateDetails().Attributes().AsRaw())
}

func TestEntityDeleteEvents(t *testing.T) {
	events, failed, err := entityDeleteEvents([]observer.Endpoint{portEndpoint}, t0)
	require.Zero(t, failed)
	require.NoError(t, err)
	require.Equal(t, 1, events.Len())

	event := events.At(0)
	assert.Equal(t, experimentalmetricmetadata.EventTypeDelete, event.EventType())
	assert.Equal(t, t0, event.Timestamp().AsTime())
	assert.Equal(t, map[string]any{
		sourcePortAttr: int64(1),
		"k8s.pod.uid":  "uid",
	}, event.ID().AsRaw())
}

func TestDeduceServiceName(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]any
		expected string
	}{
		{
			name: "service.name",
			attrs: map[string]any{
				"service.name": "my-mysql",
			},
			expected: "my-mysql",
		},
		{
			name: "k8s.pod.name",
			attrs: map[string]any{
				"k8s.pod.name": "my-daemonset-f6pxf",
				"name":         "my-name",
			},
			expected: "my-daemonset",
		},
		{
			name: "labels.app",
			attrs: map[string]any{
				"labels": map[string]any{
					"app": "my-app",
				},
			},
			expected: "my-app",
		},
		{
			name: "labels.app-name",
			attrs: map[string]any{
				"labels": map[string]any{
					"app.kubernetes.io/name": "my-app-new-name",
					"app":                    "my-app-old-name",
				},
				"process_name": "my-process",
			},
			expected: "my-app-new-name",
		},
		{
			name: "name",
			attrs: map[string]any{
				"name": "my-name",
			},
			expected: "my-name",
		},
		{
			name: "process_name",
			attrs: map[string]any{
				"process_name": "my-process",
			},
			expected: "my-process",
		},
		{
			name:     "empty",
			attrs:    map[string]any{},
			expected: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := pcommon.NewMap()
			attrs.FromRaw(tt.attrs)
			assert.Equal(t, tt.expected, deduceServiceName(attrs))
		})
	}
}
