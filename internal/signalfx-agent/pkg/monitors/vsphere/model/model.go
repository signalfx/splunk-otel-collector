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
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Host                 string `yaml:"host"`
	Port                 uint16 `yaml:"port"`
	// The vSphere username
	Username string `yaml:"username"`
	// The vSphere password
	Password string `yaml:"password"`
	// Whether we verify the server's certificate chain and host name
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
	// How often to reload the inventory and inventory metrics
	InventoryRefreshInterval timeutil.Duration `yaml:"inventoryRefreshInterval" default:"60s"`
	// Maximum number of inventory objects to be queried for performance data
	// per request. Set this value to zero (0) to request performance data for
	// all inventory objects at a time.
	PerfBatchSize int `yaml:"perfBatchSize" default:"10"`

	// An 'expr' expression to limit the inventory traversed by the monitor. Leave blank or omit
	// to traverse and get metrics for the entire vSphere inventory. Otherwise, this expression
	// is evaluated per cluster. If the expression evaluates to true, metrics are collected for
	// the objects in the cluster, otherwise it is skipped. Made available to the expr expression
	// environment are the variables: `Datacenter` and `Cluster`. For example:
	// filter: "Datacenter == 'MyDatacenter' && Cluster == 'MyCluster'"
	// The above expr value will cause metrics collection for only the given datacenter + cluster.
	// See https://github.com/antonmedv/expr for more advanced syntax.
	Filter string `yaml:"filter"`

	// The host dimension value set for monitored VMs.
	// The options are: `ip`, `hostname` and `disable`. Default is `ip`.
	// `ip`       : the VM IP if available
	// `hostname` : the VM Hostname if available
	// `disable`  : the vsphere monitor does not set the host dimension on the VM metrics
	VMHostDimension VMHostDim `yaml:"vmHostDimension" default:"ip"`

	// Path to the ca file
	TLSCACertPath string `yaml:"tlsCACertPath"`

	// Configure client certs. Both tlsClientKeyPath and tlsClientCertificatePath must be present. The files must contain
	// PEM encoded data.
	// Path to the client certificate
	TLSClientCertificatePath string `yaml:"tlsClientCertificatePath"`
	// Path to the keyfile
	TLSClientKeyPath string `yaml:"tlsClientKeyPath"`
	// When set to true, all the SOAP requests and responses will be logged.
	// This generates lots of data, only use it for debugging.
	// For this setting to take effect, make sure to restart the agent
	SOAPClientDebug bool `yaml:"soapClientDebug"`
}

type Dimensions map[string]string

type InventoryObject struct {
	Ref        types.ManagedObjectReference
	MetricIds  []types.PerfMetricId
	dimensions Dimensions
}

type Inventory struct {
	Objects      []*InventoryObject
	DimensionMap map[string]Dimensions
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
