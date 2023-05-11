package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

func TestNopFilter(t *testing.T) {
	gateway := newFakeGateway(1)
	f := nopFilter{}
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	require.NoError(t, err)
	require.Equal(t, 4, len(inv.Objects))
}

func TestExprFilterOutAllInventory(t *testing.T) {
	gateway := newFakeGateway(1)
	f, err := NewFilter("Datacenter == 'X'")
	require.NoError(t, err)
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(inv.Objects))
}

func TestExprFilterInAllInventoryInClusters(t *testing.T) {
	gateway := newFakeGateway(1)
	f, err := NewFilter("Datacenter == 'foo dc' && Cluster == 'foo cluster'")
	require.NoError(t, err)
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(inv.Objects))
}

func TestExprFilterInAllInventory(t *testing.T) {
	gateway := newFakeGateway(1)
	// there are two hosts in this DC that aren't in clusters for a total of four
	f, err := NewFilter("Datacenter == 'foo dc'")
	require.NoError(t, err)
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(inv.Objects))
}

func TestFilterInInventory(t *testing.T) {
	gateway := newFakeGateway(1)
	f, err := NewFilter("")
	require.NoError(t, err)
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	assert.NoError(t, err)
	assert.True(t, len(inv.Objects) > 0)
}

func TestFilterEmptyDC(t *testing.T) {
	gateway := newFakeGateway(1)
	f, err := NewFilter("")
	require.NoError(t, err)
	svc := NewInventorySvc(gateway, testLog, f, model.GuestIP)
	inv, err := svc.RetrieveInventory()
	assert.NoError(t, err)
	assert.True(t, len(inv.Objects) > 0)
}

func TestRetrieveInventory(t *testing.T) {
	gateway := newFakeGateway(1)
	svc := NewInventorySvc(gateway, testLog, nopFilter{}, model.GuestIP)
	inv, _ := svc.RetrieveInventory()
	requireClusterHost(t, inv, 0)
	requireClusterVM(t, inv, 1)
	requireFreeHost(t, inv, 2)
	requireFreeVM(t, inv, 3)
}

func TestVMHostnameHostDim(t *testing.T) {
	gateway := newFakeGateway(1)
	svc := NewInventorySvc(gateway, testLog, nopFilter{}, model.GuestHostName)
	inv, _ := svc.RetrieveInventory()
	dims := getDims(inv, 1)
	require.Equal(t, "foo.host.name", dims["host"])
}

func TestVMDisableHostDim(t *testing.T) {
	gateway := newFakeGateway(1)
	svc := NewInventorySvc(gateway, testLog, nopFilter{}, model.Disable)
	inv, _ := svc.RetrieveInventory()
	dims := getDims(inv, 1)
	require.Empty(t, dims["host"])
}

func requireClusterHost(t *testing.T, inv *model.Inventory, i int) {
	dims := getDims(inv, i)
	require.Equal(t, "foo cluster", dims["cluster"])
	require.Equal(t, "foo dc", dims["datacenter"])
	require.Equal(t, "4.4.4.4", dims["esx_ip"])
	require.Equal(t, "host-0", dims["ref_id"])
	require.Equal(t, model.HostType, dims["object_type"])
}

func requireClusterVM(t *testing.T, inv *model.Inventory, i int) {
	dims := getDims(inv, i)
	require.Equal(t, "foo cluster", dims["cluster"])
	require.Equal(t, "foo dc", dims["datacenter"])
	require.Equal(t, "4.4.4.4", dims["esx_ip"])
	require.Equal(t, "vm-0", dims["ref_id"])
	require.Equal(t, model.VMType, dims["object_type"])
	require.Equal(t, "foo vm", dims["vm_name"])
	require.Equal(t, "foo guest id", dims["guest_id"])
	require.Equal(t, "1.2.3.4", dims["host"])
}

func requireFreeHost(t *testing.T, inv *model.Inventory, i int) {
	dims := getDims(inv, i)
	require.Empty(t, dims["cluster"])
	require.Equal(t, "foo dc", dims["datacenter"])
	require.Equal(t, "4.4.4.4", dims["esx_ip"])
	require.Equal(t, "freehost-1", dims["ref_id"])
	require.Equal(t, model.HostType, dims["object_type"])
}

func requireFreeVM(t *testing.T, inv *model.Inventory, i int) {
	dims := getDims(inv, i)
	require.Empty(t, dims["cluster"])
	require.Equal(t, "foo dc", dims["datacenter"])
	require.Equal(t, "4.4.4.4", dims["esx_ip"])
	require.Equal(t, "vm-1", dims["ref_id"])
	require.Equal(t, model.VMType, dims["object_type"])
	require.Equal(t, "foo vm", dims["vm_name"])
	require.Equal(t, "foo guest id", dims["guest_id"])
}

func getDims(inv *model.Inventory, i int) model.Dimensions {
	return inv.DimensionMap[inv.Objects[i].Ref.Value]
}
