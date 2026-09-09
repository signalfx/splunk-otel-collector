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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	clusterConfig "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"
	networkConfig "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
	vmConfig "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/ahv/config"
	volumeConfig "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/models/volumes/v4/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configopaque"
)

func TestClusterFromV4PreservesClusterFunction(t *testing.T) {
	id := "pc-cluster"
	name := "prism-central"
	cluster := clusterFromV4(clusterConfig.Cluster{
		ExtId: &id,
		Name:  &name,
		Config: &clusterConfig.ClusterConfigReference{
			ClusterFunction: []clusterConfig.ClusterFunctionRef{clusterConfig.CLUSTERFUNCTIONREF_PRISM_CENTRAL},
		},
	})

	require.Equal(t, []string{"PRISM_CENTRAL"}, cluster.Functions)
	require.True(t, isPrismCentralCluster(cluster))
}

func TestV4InventoryConvertersPreserveCountDimensions(t *testing.T) {
	bootConfig := vmConfig.NewOneOfVmBootConfig()
	require.NoError(t, bootConfig.SetValue(*vmConfig.NewUefiBoot()))
	protectionType := vmConfig.PROTECTIONTYPE_RULE_PROTECTED
	protectionPolicyID := "policy-1"
	vm := vmFromV4(vmConfig.Vm{
		BootConfig:     bootConfig,
		Gpus:           []vmConfig.Gpu{{}},
		ProtectionType: &protectionType,
		ProtectionPolicyState: &vmConfig.ProtectionPolicyState{
			Policy: &vmConfig.PolicyReference{ExtId: &protectionPolicyID},
		},
		GuestTools: &vmConfig.GuestTools{
			IsInstalled:          boolRef(true),
			IsEnabled:            boolRef(true),
			IsReachable:          boolRef(false),
			IsVssSnapshotCapable: boolRef(true),
		},
	})
	require.Equal(t, "uefi", vm.BootType)
	require.True(t, vm.HasGPU)
	require.Equal(t, "rule_protected", vm.ProtectionType)
	require.Equal(t, "policy-1", vm.ProtectionPolicyID)
	require.Equal(t, boolRef(true), vm.GuestTools.Installed)

	encrypted := true
	container := storageContainerFromV4(clusterConfig.StorageContainer{IsEncrypted: &encrypted, ReplicationFactor: intRef(2)})
	require.Equal(t, boolRef(true), container.Encrypted)
	require.Equal(t, 2, container.ReplicationFactor)

	sharingStatus := volumeConfig.SHARINGSTATUS_SHARED
	volumeGroup := volumeGroupFromV4(volumeConfig.VolumeGroup{SharingStatus: &sharingStatus})
	require.Equal(t, "shared", volumeGroup.SharingStatus)

	storageTier := clusterConfig.STORAGETIER_SSD_MEM_NVME
	disk := diskFromV4(clusterConfig.Disk{StorageTier: &storageTier})
	require.Equal(t, "ssd_mem_nvme", disk.StorageTier)

	subnetType := networkConfig.SUBNETTYPE_OVERLAY
	subnet := subnetFromV4(networkConfig.Subnet{SubnetType: &subnetType, ClusterReference: stringRef("cluster-1"), ClusterReferenceList: []string{"cluster-2"}})
	require.Equal(t, "overlay", subnet.SubnetType)
	require.Equal(t, []string{"cluster-1", "cluster-2"}, subnet.ClusterIDs)
}

func intRef(value int) *int {
	return &value
}

func stringRef(value string) *string {
	return &value
}

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		port     int
	}{
		{
			name:     "host only",
			endpoint: "prism.example.com",
			port:     9440,
			want:     "https://prism.example.com:9440",
		},
		{
			name:     "scheme and port",
			endpoint: "http://127.0.0.1:8080",
			port:     9440,
			want:     "http://127.0.0.1:8080",
		},
		{
			name:     "strip path",
			endpoint: "https://prism.example.com:9440/foo?bar=baz",
			port:     9440,
			want:     "https://prism.example.com:9440",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEndpoint(tt.endpoint, tt.port)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestPrismV4ClientDoesNotLogHTTPRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "user" || password != "password" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(t, w, map[string]any{
			"data": []any{map[string]any{
				"extId": "cluster-1",
				"name":  "cluster-one",
			}},
			"metadata": map[string]any{"totalAvailableResults": 1},
		})
	}))
	defer server.Close()

	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = server.URL
	cfg.Username = "user"
	cfg.Password = configopaque.String("password")

	stderr := os.Stderr
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = stderr
		_ = reader.Close()
		_ = writer.Close()
	})

	client, err := newPrismV4Client(cfg)
	require.NoError(t, err)
	clusters, err := client.listClusters(context.Background())
	require.NoError(t, err)
	require.Equal(t, []nutanixCluster{{ID: "cluster-1", Name: "cluster-one"}}, clusters)

	os.Stderr = stderr
	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Empty(t, output)
}

func TestPrismV4ClientListsVMStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/vmm/v4.2/ahv/stats/vms", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("$startTime"))
		assert.Equal(t, "LAST", r.URL.Query().Get("$statType"))

		page := r.URL.Query().Get("$page")
		vmID := "vm-1"
		if page == "1" {
			vmID = "vm-2"
		}
		writeJSON(t, w, map[string]any{
			"data": []any{map[string]any{
				"extId": vmID,
				"stats": []any{
					map[string]any{"timestamp": "2026-09-09T00:00:00Z", "hypervisorNumIops": 3},
					map[string]any{"timestamp": "2026-09-09T00:00:30Z", "hypervisorNumIops": 7},
				},
			}},
			"metadata": map[string]any{"totalAvailableResults": 2},
		})
	}))
	defer server.Close()

	client := &prismClient{restClient: testV4RESTClient(t, server)}
	stats, err := client.listVMStats(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string][]metricStat{
		"vm-1": {{Name: "hypervisorNumIops", Value: 7}},
		"vm-2": {{Name: "hypervisorNumIops", Value: 7}},
	}, stats)
}
