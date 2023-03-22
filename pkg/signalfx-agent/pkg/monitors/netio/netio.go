package netio

import (
	"context"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/filter"
)

//nolint:gochecknoglobals setting net.IOCounters to a package variable for testing purposes
var iOCounters = net.IOCounters

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`
	// The network interfaces to send metrics about. This is an [overridable
	// set](https://docs.splunk.com/observability/gdi/smart-agent/smart-agent-resources.html#filtering-data-using-the-smart-agent).
	Interfaces []string `yaml:"interfaces" default:"[\"*\", \"!/^lo\\\\d*$/\", \"!/^docker.*/\", \"!/^t(un|ap)\\\\d*$/\", \"!/^veth.*$/\", \"!/^Loopback*/\"]"`
}

// structure for storing sent and received values
type netio struct {
	sent uint64
	recv uint64
}

// Monitor for Utilization
type Monitor struct {
	Output                 types.Output
	cancel                 func()
	conf                   *Config
	filter                 *filter.OverridableStringFilter
	networkTotal           uint64
	previousInterfaceStats map[string]*netio
	logger                 log.FieldLogger
}

func (m *Monitor) updateTotals(iface string, intf *net.IOCountersStat) {
	prev, ok := m.previousInterfaceStats[iface]

	// update total received
	// if there's a previous value and the counter didn't reset
	if ok && intf.BytesRecv >= prev.recv { // previous value exists and counter incremented
		m.networkTotal += (intf.BytesRecv - prev.recv)
	} else {
		// counter instance is either uninitialized or reset so add current value
		m.networkTotal += intf.BytesRecv
	}

	// update total sent
	// if there's a previous value and the counter didn't reset
	if ok && intf.BytesSent >= prev.sent {
		m.networkTotal += intf.BytesSent - prev.sent
	} else {
		// counter instance is either uninitialized or reset so add current value
		m.networkTotal += intf.BytesSent
	}

	// store values for reference next interval
	m.previousInterfaceStats[iface] = &netio{sent: intf.BytesSent, recv: intf.BytesRecv}
}

// EmitDatapoints emits a set of memory datapoints
func (m *Monitor) EmitDatapoints() {
	info, err := iOCounters(true)
	if err != nil {
		m.logger.WithError(err).Error("Failed to load net io counters")
		return
	}

	dps := make([]*datapoint.Datapoint, 0)

	for i := range info {
		intf := info[i]
		// skip it if the interface doesn't match
		if !m.filter.Matches(intf.Name) {
			m.logger.Debugf("skipping interface '%s'", intf.Name)
			continue
		}

		ifaceName := strings.Replace(intf.Name, " ", "_", -1)

		m.updateTotals(ifaceName, &intf)

		dimensions := map[string]string{"interface": ifaceName}

		dps = append(dps,
			datapoint.New("if_errors.rx", dimensions, datapoint.NewIntValue(int64(intf.Errin)), datapoint.Counter, time.Time{}),
			datapoint.New("if_errors.tx", dimensions, datapoint.NewIntValue(int64(intf.Errout)), datapoint.Counter, time.Time{}),
			datapoint.New("if_octets.rx", dimensions, datapoint.NewIntValue(int64(intf.BytesRecv)), datapoint.Counter, time.Time{}),
			datapoint.New("if_octets.tx", dimensions, datapoint.NewIntValue(int64(intf.BytesSent)), datapoint.Counter, time.Time{}),
			datapoint.New("if_packets.rx", dimensions, datapoint.NewIntValue(int64(intf.PacketsRecv)), datapoint.Counter, time.Time{}),
			datapoint.New("if_packets.tx", dimensions, datapoint.NewIntValue(int64(intf.PacketsSent)), datapoint.Counter, time.Time{}),
			datapoint.New("if_dropped.rx", dimensions, datapoint.NewIntValue(int64(intf.Dropin)), datapoint.Counter, time.Time{}),
			datapoint.New("if_dropped.tx", dimensions, datapoint.NewIntValue(int64(intf.Dropout)), datapoint.Counter, time.Time{}),
		)
	}

	// network.total
	dps = append(dps, datapoint.New("network.total", map[string]string{}, datapoint.NewIntValue(int64(m.networkTotal)), datapoint.Counter, time.Time{}))

	m.Output.SendDatapoints(dps...)
}

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	m.conf = conf

	// initialize previous stats map and network total
	m.previousInterfaceStats = map[string]*netio{}
	m.networkTotal = 0

	// configure filters
	var err error
	if len(conf.Interfaces) == 0 {
		m.filter, err = filter.NewOverridableStringFilter([]string{"*"})
		m.logger.Debugf("empty interface list, defaulting to '*'")
	} else {
		m.filter, err = filter.NewOverridableStringFilter(conf.Interfaces)
	}

	// return an error if we can't set the filter
	if err != nil {
		return err
	}

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		m.EmitDatapoints()
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
