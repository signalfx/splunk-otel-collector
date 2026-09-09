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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configopaque"
)

func TestPrismElementClientUsesV20Paths(t *testing.T) {
	responses := map[string]any{
		"/PrismGateway/services/rest/v2.0/cluster/": map[string]any{
			"id": "cluster-1", "name": "pe-cluster", "stats": map[string]any{"controller_num_iops": 42},
		},
		"/PrismGateway/services/rest/v2.0/hosts/": map[string]any{
			"entities": []any{map[string]any{"id": "host-1", "name": "host-a", "stats": map[string]any{"cpu_usage_ppm": 12}}},
		},
		"/PrismGateway/services/rest/v2.0/storage_containers/": map[string]any{
			"entities": []any{map[string]any{"id": "container-1", "name": "container-a", "usage_stats": map[string]any{"storage_usage_bytes": 100}}},
		},
		"/PrismGateway/services/rest/v2.0/vms/": map[string]any{
			"entities": []any{map[string]any{
				"id": "vm-1", "name": "vm-a", "power_state": "on", "num_vcpus": 2, "memory_mb": 1024,
				"stats": map[string]any{"hypervisor_cpu_usage_ppm": 5},
			}},
		},
		"/PrismGateway/services/rest/v2.0/volume_groups/": map[string]any{
			"entities": []any{map[string]any{"id": "vg-1", "name": "vg-a"}},
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "only GET is supported", http.StatusMethodNotAllowed)
			return
		}
		if _, _, ok := r.BasicAuth(); !ok {
			http.Error(w, "basic auth is required", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/PrismGateway/services/rest/v2.0/vms/" {
			assert.Equal(t, "true", r.URL.Query().Get("include_vm_disk_config"))
			assert.Equal(t, "true", r.URL.Query().Get("include_vm_nic_config"))
		}
		response, ok := responses[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer server.Close()

	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = server.URL
	cfg.Username = "readonly"
	cfg.Password = configopaque.String("secret")
	cfg.TLS.InsecureSkipVerify = true

	client, err := newPrismElementClient(cfg)
	require.NoError(t, err)

	clusters, err := client.listClusters(context.Background())
	require.NoError(t, err)
	require.Len(t, clusters, 1)
	require.Equal(t, "cluster-1", clusters[0].ID)
	require.Equal(t, "pe-cluster", clusters[0].Name)
	require.Equal(t, "stats_controller_num_iops", clusters[0].Stats[0].Name)

	hosts, err := client.listHosts(context.Background())
	require.NoError(t, err)
	require.Len(t, hosts, 1)
	require.Equal(t, "cluster-1", hosts[0].ClusterID)

	containers, err := client.listStorageContainers(context.Background())
	require.NoError(t, err)
	require.Len(t, containers, 1)
	require.Equal(t, "usage_stats_storage_usage_bytes", containers[0].Stats[0].Name)

	vms, err := client.listVMs(context.Background())
	require.NoError(t, err)
	require.Len(t, vms, 1)
	require.Equal(t, int64(1024*1024*1024), vms[0].MemoryBytes)
	require.Equal(t, 2, vms[0].NumCoresPerSocket)

	volumeGroups, err := client.listVolumeGroups(context.Background())
	require.NoError(t, err)
	require.Len(t, volumeGroups, 1)
}

func TestPrismElementClientFallsBackToPluralClusterPath(t *testing.T) {
	var singularRequests int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/PrismGateway/services/rest/v2.0/cluster/":
			singularRequests++
			http.NotFound(w, r)
		case "/PrismGateway/services/rest/v2.0/clusters/":
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"entities": []any{map[string]any{"uuid": "cluster-1", "name": "pe-cluster"}},
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = server.URL
	cfg.Username = "readonly"
	cfg.Password = configopaque.String("secret")
	cfg.TLS.InsecureSkipVerify = true
	client, err := newPrismElementClient(cfg)
	require.NoError(t, err)

	for range 2 {
		clusters, listErr := client.listClusters(context.Background())
		require.NoError(t, listErr)
		require.Equal(t, "cluster-1", clusters[0].ID)
	}
	require.Equal(t, 1, singularRequests, "the compatible plural path should be cached")
}
