package service

import (
	"github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/signalfx/signalfx-agent/pkg/monitors/vsphere/model"
)

// Traverses the inventory tree and returns all of the hosts and VMs.
type InventorySvc struct {
	log     logrus.FieldLogger
	gateway IGateway
	f       InvFilter
	hostDim model.VMHostDim
}

func NewInventorySvc(gateway IGateway, log logrus.FieldLogger, f InvFilter, hostDim model.VMHostDim) *InventorySvc {
	return &InventorySvc{gateway: gateway, f: f, log: log, hostDim: hostDim}
}

// use slice semantics to build parent dimensions while traversing the inv tree
type pair [2]string
type pairs []pair

const (
	dimDatacenter    = "datacenter"
	dimCluster       = "cluster"
	dimESXip         = "esx_ip"
	dimVM            = "vm_name"
	dimGuestID       = "guest_id"
	dimVMip          = "vm_ip"
	dimHost          = "host"
	dimGuestFamily   = "guest_family"
	dimGuestFullname = "guest_fullname"
)

func (svc *InventorySvc) RetrieveInventory() (*model.Inventory, error) {
	inv := model.NewInventory()
	err := svc.followFolder(inv, svc.gateway.topLevelFolderRef(), nil)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (svc *InventorySvc) followFolder(
	inv *model.Inventory,
	parentFolderRef types.ManagedObjectReference,
	dims pairs,
) error {
	var parentFolder mo.Folder
	err := svc.gateway.retrieveRefProperties(parentFolderRef, &parentFolder)
	if err != nil {
		return err
	}
	svc.debug(&parentFolder)
	for _, childRef := range parentFolder.ChildEntity {
		switch childRef.Type {
		case model.FolderType:
			err = svc.followFolder(inv, childRef, dims)
		case model.DatacenterType:
			err = svc.followDatacenter(inv, childRef, dims)
		case model.ClusterComputeType:
			err = svc.followCluster(inv, childRef, dims)
		case model.ComputeType:
			err = svc.followCompute(inv, childRef, dims)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *InventorySvc) followDatacenter(
	inv *model.Inventory,
	dcRef types.ManagedObjectReference,
	dims pairs,
) error {
	var dc mo.Datacenter
	err := svc.gateway.retrieveRefProperties(dcRef, &dc)
	if err != nil {
		return err
	}
	svc.debug(&dc)
	dims = append(dims, pair{dimDatacenter, dc.Name})
	// There is also a `dc.VmFolder` but it appears to only receive copies of VMs
	// that live under hosts. Omitting that folder to prevent double counting.
	err = svc.followFolder(inv, dc.HostFolder, dims)
	if err != nil {
		return err
	}
	return nil
}

func (svc *InventorySvc) followCluster(
	inv *model.Inventory,
	clusterRef types.ManagedObjectReference,
	dims pairs,
) error {
	var cluster mo.ClusterComputeResource
	err := svc.gateway.retrieveRefProperties(clusterRef, &cluster)
	if err != nil {
		return err
	}
	dims = append(dims, pair{dimCluster, cluster.Name})
	keep, err := svc.f.keep(dims)
	if err != nil {
		svc.log.WithError(err).Error("keep failed")
	}
	if !keep {
		svc.log.Debugf("cluster filtered: dims=[%s]", dims)
		return nil
	}
	svc.debug(&cluster)
	for _, hostRef := range cluster.ComputeResource.Host {
		err = svc.followHost(inv, hostRef, dims)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *InventorySvc) followCompute(
	inv *model.Inventory,
	computeRef types.ManagedObjectReference,
	dims pairs,
) error {
	var computeResource mo.ComputeResource
	err := svc.gateway.retrieveRefProperties(computeRef, &computeResource)
	if err != nil {
		return err
	}
	svc.debug(&computeResource)
	for _, hostRef := range computeResource.Host {
		err = svc.followHost(inv, hostRef, dims)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *InventorySvc) followHost(
	inv *model.Inventory,
	hostRef types.ManagedObjectReference,
	dims pairs,
) error {
	var host mo.HostSystem
	err := svc.gateway.retrieveRefProperties(hostRef, &host)
	if err != nil {
		return err
	}

	if host.Runtime.PowerState == types.HostSystemPowerStatePoweredOff {
		svc.log.Debugf("inventory: host powered off: name=[%s]", host.Name)
		return nil
	}

	// apply filter to hosts here in case we found hosts that aren't in a cluster
	keep, err := svc.f.keep(dims)
	if err != nil {
		svc.log.WithError(err).Error("keep failed")
	}
	if !keep {
		svc.log.Debugf("host filtered: dims=[%s]", dims)
		return nil
	}

	svc.debug(&host)
	dims = append(dims, pair{dimESXip, host.Name})
	hostDims := map[string]string{}
	amendDims(hostDims, dims)
	hostInvObj := model.NewInventoryObject(host.Self, hostDims)
	inv.AddObject(hostInvObj)
	for _, vmRef := range host.Vm {
		err = svc.followVM(inv, vmRef, dims)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc *InventorySvc) followVM(
	inv *model.Inventory,
	vmRef types.ManagedObjectReference,
	dims pairs,
) error {
	var vm mo.VirtualMachine
	err := svc.gateway.retrieveRefProperties(vmRef, &vm)
	if err != nil {
		return err
	}

	if vm.Runtime.PowerState == types.VirtualMachinePowerStatePoweredOff {
		svc.log.Debugf("inventory: vm powered off: name=[%s]", vm.Name)
		return nil
	}

	svc.debug(&vm)
	vmDims := map[string]string{
		dimVM:            vm.Name,           // e.g. "MyDebian10Host"
		dimGuestID:       vm.Config.GuestId, // e.g. "debian10_64Guest"
		dimVMip:          vm.Guest.IpAddress,
		dimGuestFamily:   vm.Guest.GuestFamily,   // e.g. "linuxGuest"
		dimGuestFullname: vm.Guest.GuestFullName, // e.g. "Other 4.x or later Linux (64-bit)"
	}

	if svc.hostDim == model.GuestIP {
		vmDims[dimHost] = vm.Guest.IpAddress
	} else if svc.hostDim == model.GuestHostName {
		vmDims[dimHost] = vm.Guest.HostName
	}

	amendDims(vmDims, dims)
	vmInvObj := model.NewInventoryObject(vm.Self, vmDims)
	inv.AddObject(vmInvObj)
	return nil
}

func (svc *InventorySvc) debug(moe mo.Entity) {
	e := moe.Entity()
	svc.log.Debugf("inventory: type=[%s] name=[%s]", e.Self.Type, e.Name)
}

func amendDims(dims map[string]string, pairs pairs) {
	for _, pair := range pairs {
		dims[pair[0]] = pair[1]
	}
}
