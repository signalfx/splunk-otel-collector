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
	"fmt"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
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
				attrs := lr.Attributes()
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutString("annotation.one", "value.one")
				annotationsMap.PutString("annotation.two", "value.two")
				attrs.PutString("endpoint", "pod.target")
				attrs.PutString("id", "pod.endpoint.id")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutString("label.one", "value.one")
				labelsMap.PutString("label.two", "value.two")
				attrs.PutString("name", "pod.name")
				attrs.PutString("namespace", "namespace")
				attrs.PutString("type", "pod")
				attrs.PutString("uid", "uid")
				lr.Body().SetStr("event.type pod endpoint pod.endpoint.id")
				return plogs
			}(),
		},
		{
			name:     "port",
			endpoint: portEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				attrs.PutString("endpoint", "port.target")
				attrs.PutString("id", "port.endpoint.id")
				attrs.PutString("name", "port.name")

				podEnvMap := attrs.PutEmptyMap("pod")
				annotationsMap := podEnvMap.PutEmptyMap("annotations")
				annotationsMap.PutString("annotation.one", "value.one")
				annotationsMap.PutString("annotation.two", "value.two")
				labelsMap := podEnvMap.PutEmptyMap("labels")
				labelsMap.PutString("label.one", "value.one")
				labelsMap.PutString("label.two", "value.two")
				podEnvMap.PutString("name", "pod.name")
				podEnvMap.PutString("namespace", "namespace")
				podEnvMap.PutString("uid", "uid")

				attrs.PutInt("port", 1)
				attrs.PutString("transport", "transport")
				attrs.PutString("type", "port")
				lr.Body().SetStr("event.type port endpoint port.endpoint.id")
				return plogs
			}(),
		},
		{
			name:     "hostport",
			endpoint: hostportEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				attrs.PutString("command", "command")
				attrs.PutString("endpoint", "hostport.target")
				attrs.PutString("id", "hostport.endpoint.id")
				attrs.PutBool("is_ipv6", true)
				attrs.PutInt("port", 1)
				attrs.PutString("process_name", "process.name")
				attrs.PutString("transport", "transport")
				attrs.PutString("type", "hostport")
				lr.Body().SetStr("event.type hostport endpoint hostport.endpoint.id")
				return plogs
			}(),
		},
		{
			name:     "container",
			endpoint: containerEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				attrs.PutInt("alternate_port", 2)
				attrs.PutString("command", "command")
				attrs.PutString("container_id", "container.id")
				attrs.PutString("endpoint", "container.target")
				attrs.PutString("host", "host")
				attrs.PutString("id", "container.endpoint.id")
				attrs.PutString("image", "image")
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutString("label.one", "value.one")
				labelsMap.PutString("label.two", "value.two")
				attrs.PutString("name", "container.name")
				attrs.PutInt("port", 1)
				attrs.PutString("tag", "tag")
				attrs.PutString("transport", "transport")
				attrs.PutString("type", "container")
				lr.Body().SetStr("event.type container endpoint container.endpoint.id")
				return plogs
			}(),
		},
		{
			name:     "k8s.node",
			endpoint: k8sNodeEndpoint,
			expectedPLogs: func() plog.Logs {
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				annotationsMap := attrs.PutEmptyMap("annotations")
				annotationsMap.PutString("annotation.one", "value.one")
				annotationsMap.PutString("annotation.two", "value.two")
				attrs.PutString("endpoint", "k8s.node.target")
				attrs.PutString("external_dns", "external.dns")
				attrs.PutString("external_ip", "external.ip")
				attrs.PutString("hostname", "host.name")
				attrs.PutString("id", "k8s.node.endpoint.id")
				attrs.PutString("internal_dns", "internal.dns")
				attrs.PutString("internal_ip", "internal.ip")
				attrs.PutInt("kubelet_endpoint_port", 1)
				labelsMap := attrs.PutEmptyMap("labels")
				labelsMap.PutString("label.one", "value.one")
				labelsMap.PutString("label.two", "value.two")
				attrs.PutString("name", "k8s.node.name")
				attrs.PutString("type", "k8s.node")
				attrs.PutString("uid", "uid")
				lr.Body().SetStr("event.type k8s.node endpoint k8s.node.endpoint.id")
				return plogs
			}(),
		},
	} {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t1 := time.Now()
			plogs, failed, err := endpointToPLogs(
				config.NewComponentIDWithName("observer.type", "observer.name"),
				"event.type", []observer.Endpoint{test.endpoint}, t0,
			)
			t2 := time.Now()
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, plogs.LogRecordCount())

			// confirm the observed timestamp is between our snapshots
			// before setting to test-friendly expected value
			lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			require.LessOrEqual(t, t1, lr.ObservedTimestamp().AsTime())
			require.GreaterOrEqual(t, t2, lr.ObservedTimestamp().AsTime())
			lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(t0))

			require.Equal(t, test.expectedPLogs, plogs, fmt.Sprintf("%s != %s", jsonify(t, test.expectedPLogs), jsonify(t, plogs)))
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
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				attrs.PutString("endpoint", "endpoint.target")
				attrs.PutString("id", "endpoint.id")
				lr.Body().SetStr("event.type endpoint endpoint.id")
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
				plogs := expectedPLogs()
				lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
				attrs := lr.Attributes()
				attrs.PutString("endpoint", "endpoint.target")
				attrs.PutString("id", "endpoint.id")
				attrs.PutString("type", "empty.details.env")
				lr.Body().SetStr("event.type empty.details.env endpoint endpoint.id")
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
				attrs := lr.Attributes()
				attrs.PutBool("annotations", false)
				attrs.PutString("endpoint", "endpoint.target")
				attrs.PutString("id", "endpoint.id")
				attrs.PutBool("labels", true)
				attrs.PutString("type", "unexpected.env")
				lr.Body().SetStr("event.type unexpected.env endpoint endpoint.id")
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
			plogs, failed, err := endpointToPLogs(
				config.NewComponentIDWithName("observer.type", "observer.name"),
				"event.type", []observer.Endpoint{test.endpoint}, t0,
			)
			if test.expectedError != "" {
				require.Error(t, err)
				require.EqualError(t, err, test.expectedError)
				return
			}
			require.NoError(t, err)
			require.Zero(t, failed)
			require.Equal(t, 1, plogs.LogRecordCount())
			lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(t0))
			require.Equal(t, test.expectedPLogs, plogs, fmt.Sprintf("%s != %s", jsonify(t, test.expectedPLogs), jsonify(t, plogs)))
		})
	}
}

func FuzzEndpointToPlogs(f *testing.F) {
	f.Add("observer.type", "observer.name", "event.type",
		"port.endpoint.id", "port.target", "port.name", "pod.name", "uid",
		"label.one", "label.value.one", "label.two", "label.value.two",
		"annotation.one", "annotation.value.one", "annotation.two", "annotation.value.two",
		"namespace", "transport", uint16(1))
	f.Add("", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", uint16(0))
	f.Fuzz(func(t *testing.T, observerType, observerName, eventType,
		endpointID, target, portName, podName, uid,
		labelOne, labelValueOne, labelTwo, labelValueTwo,
		annotationOne, annotationValueOne, annotationTwo, annotationValueTwo,
		namespace, transport string, port uint16) {
		require.NotPanics(t, func() {
			plogs, failed, err := endpointToPLogs(
				config.NewComponentIDWithName(config.Type(observerType), observerName), eventType, []observer.Endpoint{
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

			expectedLogs := expectedPLogs()
			resourceLogs := expectedLogs.ResourceLogs().At(0)
			rAttrs := resourceLogs.Resource().Attributes()
			rAttrs.PutString("discovery.event.type", eventType)
			rAttrs.PutString("discovery.observer.name", observerName)
			rAttrs.PutString("discovery.observer.type", observerType)
			expectedLR := resourceLogs.ScopeLogs().At(0).LogRecords().At(0)
			expectedLR.Body().SetStr(fmt.Sprintf("%s port endpoint %s", eventType, endpointID))
			attrs := expectedLR.Attributes()
			attrs.PutString("endpoint", target)
			attrs.PutString("id", endpointID)
			attrs.PutString("name", portName)

			podEnvMap := attrs.PutEmptyMap("pod")
			annotationsMap := podEnvMap.PutEmptyMap("annotations")
			annotationsMap.PutString(annotationOne, annotationValueOne)
			annotationsMap.PutString(annotationTwo, annotationValueTwo)
			annotationsMap.Sort()
			labelsMap := podEnvMap.PutEmptyMap("labels")
			labelsMap.PutString(labelOne, labelValueOne)
			labelsMap.PutString(labelTwo, labelValueTwo)
			labelsMap.Sort()
			podEnvMap.PutString("name", podName)
			podEnvMap.PutString("namespace", namespace)
			podEnvMap.PutString("uid", uid)
			attrs.PutInt("port", int64(port))
			attrs.PutString("transport", transport)
			attrs.PutString("type", "port")
			require.Equal(t, 1, plogs.LogRecordCount())

			lr := plogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
			lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(t0))
			require.Equal(t, expectedLogs, plogs, fmt.Sprintf("%s != %s", jsonify(t, expectedLogs), jsonify(t, plogs)))
			require.NoError(t, err)
			require.Zero(t, failed)
		})
	})
}

var (
	t0 = time.Unix(0, 0)

	logJSONMarshaler = plog.NewJSONMarshaler()

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

func expectedPLogs() plog.Logs {
	plogs := plog.NewLogs()
	rAttrs := plogs.ResourceLogs().AppendEmpty().Resource().Attributes()
	rAttrs.PutString("discovery.event.type", "event.type")
	rAttrs.PutString("discovery.observer.name", "observer.name")
	rAttrs.PutString("discovery.observer.type", "observer.type")
	sLogs := plogs.ResourceLogs().At(0).ScopeLogs().AppendEmpty()
	lr := sLogs.LogRecords().AppendEmpty()
	lr.SetTimestamp(pcommon.NewTimestampFromTime(t0))
	lr.SetObservedTimestamp(pcommon.NewTimestampFromTime(t0))
	lr.SetSeverityText("info")
	return plogs
}

func jsonify(t testing.TB, plogs plog.Logs) string {
	logBytes, err := logJSONMarshaler.MarshalLogs(plogs)
	require.NoError(t, err)
	return string(logBytes)
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
