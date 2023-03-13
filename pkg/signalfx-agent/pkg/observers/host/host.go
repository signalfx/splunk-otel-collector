// Package host observer that monitors the current host for active network
// listeners and reports them as service endpoints Use of this observer
// requires the CAP_SYS_PTRACE and CAP_DAC_READ_SEARCH capability in Linux.
package host

import (
	"fmt"
	"strconv"
	"syscall"

	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/observers"
)

const (
	observerType = "host"
)

// OBSERVER(host): Looks at the current host for listening network endpoints.
// It uses the `/proc` filesystem and requires the `SYS_PTRACE` and
// `DAC_READ_SEARCH` capabilities so that it can determine what processes own
// the listening sockets.
//
// It will look for all listening sockets on TCP and UDP over IPv4 and IPv6.

// DIMENSION(pid): The PID of the process that owns the listening endpoint

// ENDPOINT_VAR(command): The full command used to invoke this process,
// including the executable itself at the beginning.

// ENDPOINT_VAR(is_ipv6|bool): Will be `true` if the endpoint is IPv6.

// Observer that watches the current host
type Observer struct {
	serviceCallbacks *observers.ServiceCallbacks
	serviceDiffer    *observers.ServiceDiffer
	config           *Config
	logger           log.FieldLogger
}

// Config specific to the host observer
type Config struct {
	config.ObserverConfig
	// If `true`, the `pid` dimension will be omitted from the generated
	// endpoints, which means it will not appear on datapoints emitted by
	// monitors instantiated from discovery rules matching this endpoint.
	OmitPIDDimension    bool `default:"false" yaml:"omitPIDDimension"`
	PollIntervalSeconds int  `default:"10" yaml:"pollIntervalSeconds"`
}

type processName struct {
	pidNameMap map[int32]string
}

func init() {
	observers.Register(observerType, func(cbs *observers.ServiceCallbacks) interface{} {
		return &Observer{
			serviceCallbacks: cbs,
		}
	}, &Config{})
}

// Configure the host observer
func (o *Observer) Configure(config *Config) error {
	o.logger = log.WithFields(log.Fields{"observer": observerType})

	if o.serviceDiffer != nil {
		o.serviceDiffer.Stop()
	}

	o.serviceDiffer = &observers.ServiceDiffer{
		DiscoveryFn:     o.discover,
		IntervalSeconds: config.PollIntervalSeconds,
		Callbacks:       o.serviceCallbacks,
	}
	o.config = config

	o.serviceDiffer.Start()

	return nil
}

func portTypeToProtocol(t uint32) services.PortType {
	switch t {
	case syscall.SOCK_STREAM:
		return services.TCP
	case syscall.SOCK_DGRAM:
		return services.UDP
	}
	return services.UNKNOWN
}

func (o *Observer) discover() []services.Endpoint {
	conns, err := net.Connections("all")
	if err != nil {
		o.logger.WithError(err).Error("Could not get local network listeners")
		return nil
	}

	pidName := &processName{
		pidNameMap: make(map[int32]string),
	}

	err = pidName.setPidNameMap()
	if err != nil {
		o.logger.WithError(err).Error("Could not create Pid - Name Map")
		return nil
	}

	endpoints := make([]services.Endpoint, 0, len(conns))
	connsByPID := make(map[int32][]*net.ConnectionStat)
	for i := range conns {
		c := conns[i]
		// TODO: Add support for ipv6 to all observers
		isIPSocket := c.Family == syscall.AF_INET || c.Family == syscall.AF_INET6
		isTCPOrUDP := c.Type == syscall.SOCK_STREAM || c.Type == syscall.SOCK_DGRAM
		// UDP doesn't have any status
		isUDPOrListening := c.Type == syscall.SOCK_DGRAM || c.Status == "LISTEN"
		// UDP is "listening" when it has a remote port of 0
		isTCPOrHasNoRemotePort := c.Type == syscall.SOCK_STREAM || c.Raddr.Port == 0

		// PID of 0 means that the listening file descriptor couldn't be mapped
		// back to a process's set of open file descriptors in /proc
		if !isIPSocket || !isTCPOrUDP || !isUDPOrListening || !isTCPOrHasNoRemotePort || c.Pid == 0 {
			continue
		}
		connsByPID[c.Pid] = append(connsByPID[c.Pid], &c)
	}

	for pid, conns := range connsByPID {
		proc, err := process.NewProcess(pid)

		if err != nil {
			o.logger.WithFields(log.Fields{
				"pid": pid,
				"err": err,
			}).Warn("Could not examine process (it might have terminated already)")
			continue
		}

		name, err := pidName.getName(proc)
		if err != nil {
			o.logger.WithFields(log.Fields{
				"pid": pid,
				"err": err,
			}).Error("Could not get process name")
			continue
		}

		args, err := proc.Cmdline()
		if err != nil {
			o.logger.WithField("pid", pid).Error("Could not get process args")
			continue
		}

		dims := map[string]string{}
		if !o.config.OmitPIDDimension {
			dims["pid"] = strconv.Itoa(int(pid))
		}

		for _, c := range conns {
			se := services.NewEndpointCore(
				fmt.Sprintf("%s-%d-%s-%d", c.Laddr.IP, c.Laddr.Port, portTypeToProtocol(c.Type), pid), name, observerType, dims)

			se.AddExtraField("command", args)

			ip := c.Laddr.IP
			// An IP addr of 0.0.0.0 means it listens on all interfaces, including
			// localhost, so use that since we can't actually connect to 0.0.0.0.
			if ip == "0.0.0.0" {
				ip = "127.0.0.1"
			}

			se.Host = ip
			if c.Family == syscall.AF_INET6 {
				se.Host = "[" + se.Host + "]"
				se.AddExtraField("is_ipv6", true)
			} else {
				se.AddExtraField("is_ipv6", false)
			}

			se.Port = uint16(c.Laddr.Port)
			se.PortType = portTypeToProtocol(c.Type)
			se.Target = services.TargetTypeHostPort

			endpoints = append(endpoints, se)
		}
	}
	return endpoints
}

// Shutdown the service differ routine
func (o *Observer) Shutdown() {
	if o.serviceDiffer != nil {
		o.serviceDiffer.Stop()
	}
}
