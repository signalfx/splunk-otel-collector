package mssqlserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/winperfcounters"

	"github.com/influxdata/telegraf"
	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/sqlserver"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	Host                 string `yaml:"host" validate:"required" default:"."`
	Port                 uint16 `yaml:"port" validate:"required" default:"1433"`
	// UserID used to access the SQL Server instance.
	UserID string `yaml:"userID"`
	// Password used to access the SQL Server instance.
	Password string `yaml:"password" neverLog:"true"`
	// The app name used by the monitor when connecting to the SQLServer.
	AppName string `yaml:"appName" default:"signalfxagent"`
	// The version of queries to use when accessing the cluster.
	// Please refer to the telegraf documentation for more information.
	QueryVersion int `yaml:"queryVersion" default:"2"`
	// Whether the database is an azure database or not.
	AzureDB bool `yaml:"azureDB"`
	// Queries to exclude possible values are `PerformanceCounters`, `WaitStatsCategorized`,
	// `DatabaseIO`, `DatabaseProperties`, `CPUHistory`, `DatabaseSize`, `DatabaseStats`, `MemoryClerk`
	// `VolumeSpace`, and `PerformanceMetrics`.
	ExcludeQuery []string `yaml:"excludedQueries"`
	// Log level to use when accessing the database
	Log uint `yaml:"log" default:"1"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel func()
	logger log.FieldLogger
}

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	plugin := telegrafInputs.Inputs["sqlserver"]().(*telegrafPlugin.SQLServer)

	server := fmt.Sprintf("Server=%s;Port=%d;", conf.Host, conf.Port)

	if conf.UserID != "" {
		server = fmt.Sprintf("%sUser Id=%s;", server, conf.UserID)
	}
	if conf.Password != "" {
		server = fmt.Sprintf("%sPassword=%s;", server, conf.Password)
	}
	if conf.AppName != "" {
		server = fmt.Sprintf("%sapp name=%s;", server, conf.AppName)
	}
	server = fmt.Sprintf("%slog=%d;", server, conf.Log)

	plugin.Servers = []string{server}
	plugin.QueryVersion = conf.QueryVersion
	plugin.AzureDB = conf.AzureDB
	plugin.ExcludeQuery = conf.ExcludeQuery

	// create batch emitter
	emit := baseemitter.NewEmitter(m.Output, m.logger)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	emit.AddTag("plugin", strings.Replace(monitorType, "/", "-", -1))

	// replacer sanitizes metrics according to our PCR reporter rules (ours have to come first).
	replacer := strings.NewReplacer(append([]string{"%", "pct", "(s)", "_"}, winperfcounters.MetricReplacements...)...)

	emit.AddMetricNameTransformation(func(metric string) string {
		return strings.Trim(replacer.Replace(strings.ToLower(metric)), "_")
	})

	emit.AddMeasurementTransformation(
		func(ms telegraf.Metric) error {
			// if it's a sqlserver_performance metric
			// remap the counter and value to a field
			if ms.Name() == "sqlserver_performance" {
				emitter.RenameFieldWithTag(ms, "counter", "value", replacer)
			}

			// if it's a sqlserver_memory_clerks metric remap clerk type to field
			if ms.Name() == "sqlserver_memory_clerks" {
				ms.SetName(fmt.Sprintf("sqlserver_memory_clerks.size_kb"))
				emitter.RenameFieldWithTag(ms, "clerk_type", "size_kb", replacer)
			}
			return nil
		})

	// convert the metric name to lower case
	emit.AddMetricNameTransformation(strings.ToLower)

	// create the accumulator
	ac := accumulator.NewAccumulator(emit)

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics from the plugin")
		}

	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
