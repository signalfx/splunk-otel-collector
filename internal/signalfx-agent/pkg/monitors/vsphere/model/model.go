package model

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
	"github.com/vmware/govmomi/vim25/types"
)

// "real-time" vsphereInfo metrics are available in 20 second intervals
const RealtimeMetricsInterval = 20

const (
	DatacenterType     = "Datacenter"
	ClusterComputeType = "ClusterComputeResource"
	ComputeType        = "ComputeResource"
	VMType             = "VirtualMachine"
	HostType           = "HostSystem"
	FolderType         = "Folder"
)

type VMHostDim string

const (
	GuestIP       VMHostDim = "ip"
	GuestHostName VMHostDim = "hostname"
	Disable       VMHostDim = "disable"
)

// Config for the vSphere monitor
type Config struct {
	config.MonitorConfig     `yaml:",inline" acceptsEndpoints:"true"`
	VMHostDimension          VMHostDim         `yaml:"vmHostDimension" default:"ip"`
	Username                 string            `yaml:"username"`
	Password                 string            `yaml:"password"`
	Filter                   string            `yaml:"filter"`
	Host                     string            `yaml:"host"`
	TLSCACertPath            string            `yaml:"tlsCACertPath"`
	TLSClientCertificatePath string            `yaml:"tlsClientCertificatePath"`
	TLSClientKeyPath         string            `yaml:"tlsClientKeyPath"`
	InventoryRefreshInterval timeutil.Duration `yaml:"inventoryRefreshInterval" default:"60s"`
	PerfBatchSize            int               `yaml:"perfBatchSize" default:"10"`
	Port                     uint16            `yaml:"port"`
	InsecureSkipVerify       bool              `yaml:"insecureSkipVerify"`
	SOAPClientDebug          bool              `yaml:"soapClientDebug"`
}

type Dimensions map[string]string

type InventoryObject struct {
	dimensions Dimensions
	Ref        types.ManagedObjectReference
	MetricIds  []types.PerfMetricId
}

type Inventory struct {
	DimensionMap map[string]Dimensions
	Objects      []*InventoryObject
}

// Validate that VMHostDimension is one of the supported options
func (c *Config) Validate() error {
	switch c.VMHostDimension {
	case GuestIP, GuestHostName, Disable:
		return nil
	default:
		return fmt.Errorf("hostDimensionValue '%s' is invalid, it can only be '%s', '%s' or '%s'", c.VMHostDimension, GuestIP, GuestHostName, Disable)
	}
}

func NewInventoryObject(ref types.ManagedObjectReference, extraDimensions map[string]string) *InventoryObject {
	dimensions := map[string]string{
		"ref_id":      ref.Value,
		"object_type": ref.Type,
	}
	for key, value := range extraDimensions {
		dimensions[key] = value
	}
	return &InventoryObject{
		Ref:        ref,
		dimensions: dimensions,
	}
}

func NewInventory() *Inventory {
	inv := &Inventory{}
	inv.DimensionMap = make(map[string]Dimensions)
	return inv
}

func (inv *Inventory) AddObject(obj *InventoryObject) {
	inv.Objects = append(inv.Objects, obj)
	inv.DimensionMap[obj.Ref.Value] = obj.dimensions
}

type MetricInfosByKey map[int32]MetricInfo

type MetricInfo struct {
	MetricName      string
	PerfCounterInfo types.PerfCounterInfo
}

type VsphereInfo struct {
	Inv              *Inventory
	PerfCounterIndex MetricInfosByKey
}
