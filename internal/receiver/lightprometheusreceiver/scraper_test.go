// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lightprometheusreceiver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	conventions "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func TestScraper(t *testing.T) {
	promMock := newPromMockServer(t)
	u, err := url.Parse(promMock.URL)
	require.NoError(t, err)

	tests := []struct {
		cfg                        *Config
		expectedResourceAttributes map[string]any
		name                       string
	}{
		{
			name: "default_config",
			cfg:  createDefaultConfig().(*Config),
			expectedResourceAttributes: map[string]any{
				string(conventions.ServiceNameKey):       "",
				string(conventions.ServiceInstanceIDKey): u.Host,
			},
		},
		{
			name: "all_resource_attributes",
			cfg: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.ResourceAttributes.ServiceName.Enabled = true
				cfg.ResourceAttributes.URLScheme.Enabled = true
				cfg.ResourceAttributes.ServerPort.Enabled = true
				cfg.ResourceAttributes.ServerAddress.Enabled = true
				return cfg
			}(),
			expectedResourceAttributes: map[string]any{
				string(conventions.ServiceNameKey):       "",
				string(conventions.ServiceInstanceIDKey): u.Host,
				string(conventions.ServerAddressKey):     u.Host,
				string(conventions.ServerPortKey):        u.Port(),
				string(conventions.URLSchemeKey):         "http",
			},
		},
		{
			name: "deprecated_resource_attributes",
			cfg: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.ResourceAttributes.ServiceName.Enabled = true
				cfg.ResourceAttributes.HTTPScheme.Enabled = true
				cfg.ResourceAttributes.NetHostPort.Enabled = true
				cfg.ResourceAttributes.NetHostName.Enabled = true
				return cfg
			}(),
			expectedResourceAttributes: map[string]any{
				string(conventions.ServiceNameKey):       "",
				string(conventions.ServiceInstanceIDKey): u.Host,
				string(conventions.NetHostNameKey):       u.Host,
				string(conventions.NetHostPortKey):       u.Port(),
				string(conventions.HTTPSchemeKey):        "http",
			},
		},
		{
			name: "no_resource_attributes",
			cfg: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.ResourceAttributes.ServiceInstanceID.Enabled = false
				cfg.ResourceAttributes.ServiceName.Enabled = false
				return cfg
			}(),
			expectedResourceAttributes: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			cfg.ClientConfig.Endpoint = fmt.Sprintf("%s%s", promMock.URL, "/metrics")
			require.NoError(t, xconfmap.Validate(cfg))

			scraper := newScraper(receivertest.NewNopSettings(receivertest.NopType), cfg)

			err := scraper.start(context.Background(), componenttest.NewNopHost())
			require.NoError(t, err)

			actualMetrics, err := scraper.scrape(context.Background())
			require.NoError(t, err)

			expectedFile := filepath.Join("testdata", "scraper", "expected.json")
			expectedMetrics, err := readMetrics(expectedFile)
			require.NoError(t, err)
			require.NoError(t, expectedMetrics.ResourceMetrics().At(0).Resource().Attributes().FromRaw(tt.expectedResourceAttributes))

			require.NoError(t, pmetrictest.CompareMetrics(expectedMetrics, actualMetrics,
				pmetrictest.IgnoreResourceAttributeValue("service.name"),
				pmetrictest.IgnoreMetricDataPointsOrder(), pmetrictest.IgnoreStartTimestamp(),
				pmetrictest.IgnoreTimestamp(), pmetrictest.IgnoreMetricsOrder()))
		})
	}
}

func readMetrics(filePath string) (pmetric.Metrics, error) {
	expectedFileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return pmetric.Metrics{}, err
	}
	unmarshaller := &pmetric.JSONUnmarshaler{}
	return unmarshaller.UnmarshalMetrics(expectedFileBytes)
}

func newPromMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/metrics" {
			rw.WriteHeader(200)
			_, err := rw.Write([]byte(`# HELP istio_agent_cert_expiry_seconds The time remaining, in seconds, before the certificate chain will expire. A negative    value indicates the cert is expired.
# TYPE istio_agent_cert_expiry_seconds gauge
istio_agent_cert_expiry_seconds{resource_name="default"} 7449418.02275500029
# HELP istio_agent_endpoint_no_pod Endpoints without an associated pod.
# TYPE istio_agent_endpoint_no_pod gauge
istio_agent_endpoint_no_pod 0
# HELP istio_agent_go_gc_cycles_automatic_gc_cycles_total Count of completed GC cycles generated by the Go runtime.
# TYPE istio_agent_go_gc_cycles_automatic_gc_cycles_total counter
istio_agent_go_gc_cycles_automatic_gc_cycles_total 5355
# HELP istio_agent_go_gc_cycles_forced_gc_cycles_total Count of completed GC cycles forced by the application.
# TYPE istio_agent_go_gc_cycles_forced_gc_cycles_total counter
istio_agent_go_gc_cycles_forced_gc_cycles_total 0
# HELP istio_agent_go_gc_cycles_total_gc_cycles_total Count of all completed GC cycles.
# TYPE istio_agent_go_gc_cycles_total_gc_cycles_total counter
istio_agent_go_gc_cycles_total_gc_cycles_total 5355
# HELP istio_agent_go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE istio_agent_go_gc_duration_seconds summary
istio_agent_go_gc_duration_seconds{quantile="0"} 7.6654e-05
istio_agent_go_gc_duration_seconds{quantile="0.25"} 0.000134994
istio_agent_go_gc_duration_seconds{quantile="0.5"} 0.000204339
istio_agent_go_gc_duration_seconds{quantile="0.75"} 0.000471661
istio_agent_go_gc_duration_seconds{quantile="1"} 0.097528526
istio_agent_go_gc_duration_seconds_sum 3.420650096
istio_agent_go_gc_duration_seconds_count 5355
# HELP istio_agent_go_gc_heap_allocs_by_size_bytes_total Distribution of heap allocations by approximate size. Note that this  does not include tiny objects as defined by /gc/heap/tiny/allocs:objects, only tiny blocks.
# TYPE istio_agent_go_gc_heap_allocs_by_size_bytes_total histogram
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="8.999999999999998"} 69482
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="24.999999999999996"} 3.331236e+06
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="64.99999999999999"} 3.4379917e+07
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="144.99999999999997"} 1.57242824e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="320.99999999999994"} 1.74101606e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="704.9999999999999"} 1.89289766e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="1536.9999999999998"} 2.18787074e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="3200.9999999999995"} 2.18855996e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="6528.999999999999"} 2.18891888e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="13568.999999999998"} 2.18921794e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="27264.999999999996"} 2.19043605e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_bucket{le="+Inf"} 2.19116234e+08
istio_agent_go_gc_heap_allocs_by_size_bytes_total_sum 3.10449604208e+11
istio_agent_go_gc_heap_allocs_by_size_bytes_total_count 2.19116234e+08
# TYPE istio_requests_total counter
istio_requests_total{response_code="foo-bar-value",reporter="foo-bar-value",source_workload="foo-bar-value",source_workload_namespace="foo-bar-value",source_principal="foo-bar-value",source_app="foo-bar-value",source_version="foo-bar-value",source_cluster="foo-bar-value",destination_workload="foo-bar-value",destination_workload_namespace="foo-bar-value",destination_principal="foo-bar-value",destination_app="foo-bar-value",destination_version="foo-bar-value",destination_service="foo-bar-value",destination_service_name="foo-bar-value",destination_service_namespace="foo-bar-value",destination_cluster="foo-bar-value",request_protocol="foo-bar-value",response_flags="foo-bar-value",grpc_response_status="foo-bar-value",connection_security_policy="foo-bar-value",source_canonical_service="foo-bar-value",destination_canonical_service="foo-bar-value",source_canonical_revision="foo-bar-value",destination_canonical_revision="foo-bar-value"} 1
istio_requests_total{response_code="foo-bar-value",reporter="foo-bar-value",source_workload="foo-bar-value",source_workload_namespace="foo-bar-value",source_principal="foo-bar-value",source_app="foo-bar-value",source_version="foo-bar-value",source_cluster="foo-bar-value",destination_workload="foo-bar-value",destination_workload_namespace="foo-bar-value",destination_principal="foo-bar-value",destination_app="foo-bar-value",destination_version="foo-bar-value",destination_service="foo-bar-value",destination_service_name="foo-bar-value",destination_service_namespace="foo-bar-value",destination_cluster="foo-bar-value",request_protocol="foo-bar-value",response_flags="foo-bar-value",grpc_response_status="foo-bar-value",connection_security_policy="foo-bar-value",source_canonical_service="foo-bar-value",destination_canonical_service="foo-bar-value",source_canonical_revision="foo-bar-value",destination_canonical_revision="foo-bar-value"} 2
istio_requests_total{response_code="foo-bar-value",reporter="foo-bar-value",source_workload="foo-bar-value",source_workload_namespace="foo-bar-value",source_principal="foo-bar-value",source_app="foo-bar-value",source_version="foo-bar-value",source_cluster="foo-bar-value",destination_workload="foo-bar-value",destination_workload_namespace="foo-bar-value",destination_principal="foo-bar-value",destination_app="foo-bar-value",destination_version="foo-bar-value",destination_service="foo-bar-value",destination_service_name="foo-bar-value",destination_service_namespace="foo-bar-value",destination_cluster="foo-bar-value",request_protocol="foo-bar-value",response_flags="foo-bar-value",grpc_response_status="foo-bar-value",connection_security_policy="foo-bar-value",source_canonical_service="foo-bar-value",destination_canonical_service="foo-bar-value",source_canonical_revision="foo-bar-value",destination_canonical_revision="foo-bar-value"} 1
# TYPE istio_request_bytes histogram
istio_request_bytes_bucket{response_code="200",reporter="foo-bar-value",source_workload="foo-bar-value",source_workload_namespace="foo-bar-value",source_principal="foo-bar-value",source_app="foo-bar-value",source_version="foo-bar-value",source_cluster="foo-bar-value",destination_workload="foo-bar-value",destination_workload_namespace="foo-bar-value",destination_principal="foo-bar-value",destination_app="foo-bar-value",destination_version="foo-bar-value",destination_service="foo-bar-value",destination_service_name="foo-bar-value",destination_service_namespace="foo-bar-value",destination_cluster="foo-bar-value",request_protocol="foo-bar-value",response_flags="foo-bar-value",grpc_response_status="foo-bar-value",connection_security_policy="unknown",source_canonical_service="foo-bar-value",destination_canonical_service="foo-bar-value",source_canonical_revision="foo-bar-value",destination_canonical_revision="foo-bar-value",le="+Inf"} 0
`))
			require.NoError(t, err)
			return
		}
		rw.WriteHeader(404)
	}))
}
