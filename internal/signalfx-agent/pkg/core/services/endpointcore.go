package services

import (
	"fmt"
	"regexp"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// TargetType represents the type of the resource an endpoint refers to.
type TargetType string

const (
	TargetTypePod            TargetType = "pod"
	TargetTypeHostPort       TargetType = "hostport"
	TargetTypeContainer      TargetType = "container"
	TargetTypeKubernetesNode TargetType = "k8s-node"
)

// PortType represents the transport protocol used to communicate with this port
type PortType string

const (
	UDP     PortType = "UDP"
	TCP     PortType = "TCP"
	UNKNOWN PortType = "UNKNOWN"
)

//nolint:gochecknoglobals
var ipAddrRegexp = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

var _ config.CustomConfigurable = &EndpointCore{}

// EndpointCore represents an exposed network target.  This target can be a
// full host/port endpoint or simply a hostname/ip.
type EndpointCore struct {
	ID ID `yaml:"id"`
	// A observer assigned name of the endpoint. For example, if using the
	// `k8s-api` observer, `name` will be the port name in the pod spec, if
	// any.
	Name string `yaml:"name"`
	// The hostname/IP address of the endpoint.  If this is an IPv6 address, it
	// will be surrounded by `[` and `]`.
	Host string `yaml:"host"`
	// TCP or UDP
	PortType PortType `yaml:"port_type"`
	// The TCP/UDP port number of the endpoint
	Port uint16 `yaml:"port"`
	// The type of the thing that this endpoint directly refers to.  If the
	// endpoint has a host and port associated with it (most common), the value
	// will be `hostport`.  Other possible values are: `pod`, `container`,
	// `host`.  See the docs for the specific observer you are using for more
	// details on what types that observer emits.
	Target TargetType `yaml:"target"`
	// The observer that discovered this endpoint
	DiscoveredBy  string                 `yaml:"discovered_by"`
	Configuration map[string]interface{} `yaml:"-"`
	// The type of monitor that this endpoint has requested.  This is populated
	// by observers that pull configuration directly from the platform they are
	// observing.
	MonitorType     string                 `yaml:"-"`
	extraDimensions map[string]string      `yaml:"-"`
	extraFields     map[string]interface{} `yaml:"-"`
}

// NewEndpointCore returns a new initialized endpoint core struct
func NewEndpointCore(id string, name string, discoveredBy string, dims map[string]string) *EndpointCore {
	if id == "" {
		// Observers must provide an ID or else they are majorly broken
		panic("EndpointCore cannot be created without an id")
	}

	ec := &EndpointCore{
		ID:              ID(id),
		Name:            name,
		DiscoveredBy:    discoveredBy,
		extraDimensions: dims,
		extraFields:     map[string]interface{}{},
	}

	return ec
}

func (e *EndpointCore) String() string {
	return fmt.Sprintf("id: %s; name: %s; host: %s; port-type: %s; port: %d; target: %s; discovered-by: %s", e.ID, e.Name, e.Host, e.PortType, e.Port, e.Target, e.DiscoveredBy)
}

// Core returns the EndpointCore since it will be embedded in an Endpoint
// instance
func (e *EndpointCore) Core() *EndpointCore {
	return e
}

// ENDPOINT_VAR(network_port): An alias for `port`

// ENDPOINT_VAR(ip_address): The IP address of the endpoint if the `host` is in
// the from of an IPv4 address

// ENDPOINT_VAR(has_port): Set to `true` if the endpoint has a port assigned to
// it.  This will be `false` for endpoints that represent a host/container as a
// whole.

// DerivedFields returns aliased and computed variable fields for this endpoint
func (e *EndpointCore) DerivedFields() map[string]interface{} {
	out := map[string]interface{}{
		"network_port": e.Port,
	}
	if ipAddrRegexp.MatchString(e.Host) {
		out["ip_address"] = e.Host
	}
	out["has_port"] = e.Port != 0

	return utils.MergeInterfaceMaps(utils.CloneInterfaceMap(e.extraFields), utils.StringMapToInterfaceMap(e.Dimensions()), out)
}

// ExtraConfig returns a map of values to be considered when configuring a
// monitor.  These values will take precedence over anything the user
// specifies.
func (e *EndpointCore) ExtraConfig() (map[string]interface{}, error) {
	vars := map[string]interface{}{
		"name": utils.FirstNonEmpty(e.Name, string(e.ID)),
	}

	if e.Host != "" {
		vars["host"] = e.Host
	}

	// Port 0 is never valid and so it means that the port is unknown or not
	// applicable to this endpoint.
	if e.Port != 0 {
		vars["port"] = e.Port
	}

	return utils.MergeInterfaceMaps(vars, e.Configuration), nil
}

// IsSelfConfigured tells whether this endpoint comes with enough configuration
// to run without being configured further.  This ultimately just means whether
// it specifies what type of monitor to use to monitor it.
func (e *EndpointCore) IsSelfConfigured() bool {
	return e.MonitorType != ""
}

// Dimensions returns a map of dimensions set on this endpoint
func (e *EndpointCore) Dimensions() map[string]string {
	return utils.CloneStringMap(e.extraDimensions)
}

// AddDimension adds a dimension to this endpoint
func (e *EndpointCore) AddDimension(k string, v string) {
	if e.extraDimensions == nil {
		e.extraDimensions = make(map[string]string)
	}

	e.extraDimensions[k] = v
}

// RemoveDimension removes a dimension from this endpoint
func (e *EndpointCore) RemoveDimension(k string) {
	delete(e.extraDimensions, k)
}

func (e *EndpointCore) AddExtraField(name string, val interface{}) {
	e.extraFields[name] = val
}
