// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package nutanixreceiver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV4RESTClientListAllPaginates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/test/v4.2/config/entities", r.URL.Path)
		page, err := strconv.Atoi(r.URL.Query().Get("$page"))
		assert.NoError(t, err)
		count := v4PageSize
		if page == 1 {
			count = 1
		}
		entities := make([]map[string]any, count)
		for i := range entities {
			entities[i] = map[string]any{"extId": strconv.Itoa(page*v4PageSize + i)}
		}
		writeJSON(t, w, map[string]any{
			"data":     entities,
			"metadata": map[string]any{"totalAvailableResults": 101},
		})
	}))
	defer server.Close()

	entities, err := testV4RESTClient(t, server).listAll(context.Background(), "/api/test/v4.2/config/entities")
	require.NoError(t, err)
	require.Len(t, entities, 101)
}

func TestV4RESTClientRetriesServerErrors(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) < 3 {
			http.Error(w, "retry", http.StatusServiceUnavailable)
			return
		}
		writeJSON(t, w, map[string]any{"data": []any{}, "metadata": map[string]any{"totalAvailableResults": 0}})
	}))
	defer server.Close()

	count, err := testV4RESTClient(t, server).count(context.Background(), "/api/test")
	require.NoError(t, err)
	require.Zero(t, count)
	require.Equal(t, int32(3), requests.Load())
}

func TestCollectPrismCentralContinuesAfterEndpointFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/prism/v4.2/config/categories":
			http.NotFound(w, r)
		case "/api/prism/v4.2/config/tasks":
			total := 2
			if r.URL.Query().Get("$filter") != "" {
				total = 1
			}
			writeJSON(t, w, map[string]any{
				"data":     []any{map[string]any{"status": "RUNNING"}},
				"metadata": map[string]any{"totalAvailableResults": total},
			})
		case "/api/monitoring/v4.2/serviceability/alerts":
			writeJSON(t, w, map[string]any{
				"data":     []any{map[string]any{"severity": "CRITICAL", "isResolved": false, "isAcknowledged": true}},
				"metadata": map[string]any{"totalAvailableResults": 1},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	snapshot := testV4RESTClient(t, server).collectPrismCentral(context.Background())
	require.Len(t, snapshot.Errors, 1)
	require.InDelta(t, 1, additionalMetricValue(t, snapshot.Metrics, "nutanix.prism.entity.count", map[string]string{
		"nutanix.entity.type":       "task",
		"nutanix.entity.state_type": "status",
		"nutanix.entity.state":      "running",
	}), 0.001)
	require.InDelta(t, 1, additionalMetricValue(t, snapshot.Metrics, "nutanix.monitoring.entity.count", map[string]string{
		"nutanix.entity.type":       "alert",
		"nutanix.entity.state_type": "unresolved_severity",
		"nutanix.entity.state":      "critical",
	}), 0.001)
}

func TestV4RESTClientStatsExtractsLatestValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.URL.Query().Get("$startTime"))
		assert.Equal(t, "LAST", r.URL.Query().Get("$statType"))
		writeJSON(t, w, map[string]any{
			"data": map[string]any{
				"numberOfOperations": []any{
					map[string]any{"timestamp": "2026-09-08T00:00:00Z", "value": 3},
					map[string]any{"timestamp": "2026-09-08T00:00:30Z", "value": 7},
				},
			},
		})
	}))
	defer server.Close()

	stats, err := testV4RESTClient(t, server).stats(context.Background(), "/api/test/stats/entity")
	require.NoError(t, err)
	require.Equal(t, []metricStat{{Name: "numberOfOperations", Value: 7}}, stats)
}

func TestCollectAdditionalMetricsDisabledMakesNoRequests(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()

	client := &prismClient{restClient: testV4RESTClient(t, server)}
	snapshot, err := client.collectAdditionalMetrics(context.Background(), additionalMetricsRequest{})
	require.NoError(t, err)
	require.Empty(t, snapshot.Metrics)
	require.Empty(t, snapshot.Errors)
	require.Zero(t, requests.Load())
}

func TestCollectAdditionalMetricsCoversEveryOptionalDomain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, map[string]any{
			"data":     []any{},
			"metadata": map[string]any{"totalAvailableResults": 0},
		})
	}))
	defer server.Close()

	client := &prismClient{restClient: testV4RESTClient(t, server)}
	snapshot, err := client.collectAdditionalMetrics(context.Background(), additionalMetricsRequest{
		DataProtection:    true,
		Files:             true,
		Microsegmentation: true,
		Networking:        true,
		Objects:           true,
		PrismCentral:      true,
	})
	require.NoError(t, err)
	require.Empty(t, snapshot.Errors)

	metricNames := map[string]bool{}
	for _, metric := range snapshot.Metrics {
		metricNames[metric.Name] = true
	}
	for _, name := range []string{
		"nutanix.data_protection.entity.count",
		"nutanix.files.entity.count",
		"nutanix.microseg.entity.count",
		"nutanix.monitoring.entity.count",
		"nutanix.networking.entity.count",
		"nutanix.objects.entity.count",
		"nutanix.prism.entity.count",
	} {
		require.Truef(t, metricNames[name], "expected %s", name)
	}
}

func TestCountProtectionPolicySchedules(t *testing.T) {
	policies := []map[string]any{
		{
			"replicationConfigurations": []any{
				map[string]any{"schedule": map[string]any{"recoveryPointType": "CRASH_CONSISTENT", "recoveryPointObjectiveTimeSeconds": 0}},
				map[string]any{"schedule": map[string]any{"recoveryPointType": "CRASH_CONSISTENT", "recoveryPointObjectiveTimeSeconds": 0}},
				map[string]any{"schedule": map[string]any{"recoveryPointType": "APPLICATION_CONSISTENT", "recoveryPointObjectiveTimeSeconds": 3600}},
				map[string]any{"schedule": map[string]any{"recoveryPointType": "APPLICATION_CONSISTENT", "recoveryPointObjectiveTimeSeconds": 3600}},
			},
		},
	}

	counts := countProtectionPolicySchedules(policies)
	require.Equal(t, 2, counts["total"])
	require.Equal(t, 1, counts["crash_consistent"])
	require.Equal(t, 1, counts["application_consistent"])
	require.Equal(t, 1, counts["sync"])
	require.Equal(t, 1, counts["async"])
}

func TestCountProtectionPolicySchedulesIgnoresMissingRPO(t *testing.T) {
	counts := countProtectionPolicySchedules([]map[string]any{{
		"replicationConfigurations": []any{
			map[string]any{"schedule": map[string]any{"recoveryPointType": "CRASH_CONSISTENT"}},
		},
	}})

	require.Equal(t, 1, counts["total"])
	require.Equal(t, 1, counts["crash_consistent"])
	require.Zero(t, counts["sync"])
}

func TestCountProtectedVMsByRPO(t *testing.T) {
	policies := []map[string]any{
		{
			"extId": "sync-policy",
			"replicationConfigurations": []any{
				map[string]any{"schedule": map[string]any{"recoveryPointObjectiveTimeSeconds": 0}},
			},
		},
		{
			"extId": "nearsync-policy",
			"replicationConfigurations": []any{
				map[string]any{"schedule": map[string]any{"recoveryPointObjectiveTimeSeconds": 600}},
			},
		},
		{
			"extId": "async-policy",
			"replicationConfigurations": []any{
				map[string]any{"schedule": map[string]any{"recoveryPointObjectiveTimeSeconds": 3600}},
			},
		},
	}
	vms := []nutanixVM{
		{ProtectionPolicyID: "sync-policy"},
		{ProtectionPolicyID: "sync-policy"},
		{ProtectionPolicyID: "nearsync-policy"},
		{ProtectionPolicyID: "async-policy"},
		{ProtectionPolicyID: "missing-policy"},
	}

	require.Equal(t, map[string]int{"sync": 2, "nearsync": 1, "async": 1}, countProtectedVMsByRPO(policies, vms))
}

func TestMicrosegmentationScopeGroupsAPIEnums(t *testing.T) {
	policies := []map[string]any{
		{"scope": "ALL_VLAN"},
		{"scope": "ALL_VPC"},
		{"scope": "VPC_LIST"},
	}

	require.Equal(t, 1, countMicrosegPoliciesByScope(policies, "vlan", "all_vlan"))
	require.Equal(t, 2, countMicrosegPoliciesByScope(policies, "vpc", "all_vpc", "vpc_list"))
}

func testV4RESTClient(t *testing.T, server *httptest.Server) *v4RESTClient {
	t.Helper()
	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	return &v4RESTClient{
		baseURL:    baseURL,
		httpClient: server.Client(),
		interval:   30 * time.Second,
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(value))
}

func additionalMetricValue(t *testing.T, metrics []additionalMetric, name string, attrs map[string]string) float64 {
	t.Helper()
	for _, metric := range metrics {
		if metric.Name != name {
			continue
		}
		matches := true
		for key, value := range attrs {
			if metric.Attributes[key] != value {
				matches = false
				break
			}
		}
		if matches {
			return metric.Value
		}
	}
	t.Fatalf("additional metric %q with attributes %v not found", name, attrs)
	return 0
}
