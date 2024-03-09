package telegrafsnmp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ulule/deepcopier"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/snmp"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Field represents an SNMP field
type Field struct {
	Name           string `yaml:"name"`
	Oid            string `yaml:"oid"`
	OidIndexSuffix string `yaml:"oidIndexSuffix"`
	Conversion     string `yaml:"conversion"`
	OidIndexLength int    `yaml:"oidIndexLength"`
	IsTag          bool   `yaml:"isTag"`
}

// Table represents an SNMP table
type Table struct {
	Name        string   `yaml:"name"`
	Oid         string   `yaml:"oid"`
	InheritTags []string `yaml:"inheritTags"`
	Fields      []Field  `yaml:"field"`
	IndexAsTag  bool     `yaml:"indexAsTag"`
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	EngineID             string   `yaml:"engineID"`
	Name                 string   `yaml:"name"`
	SecLevel             string   `yaml:"secLevel" default:"noAuthNoPriv"`
	SecName              string   `yaml:"secName"`
	Host                 string   `yaml:"host"`
	Community            string   `yaml:"community" default:"public"`
	PrivPassword         string   `yaml:"privPassword"`
	AuthProtocol         string   `yaml:"authProtocol" default:""`
	PrivProtocol         string   `yaml:"privProtocol" default:""`
	AuthPassword         string   `yaml:"authPassword" default:"" neverLog:"true"`
	ContextName          string   `yaml:"contextName"`
	Tables               []Table  `yaml:"tables"`
	Fields               []Field  `yaml:"fields"`
	Agents               []string `yaml:"agents"`
	Retries              int      `yaml:"retries"`
	EngineBoots          uint32   `yaml:"engineBoots"`
	EngineTime           uint32   `yaml:"engineTime"`
	Port                 uint16   `yaml:"port"`
	MaxRepetitions       uint8    `yaml:"maxRepetitions" default:"50"`
	Version              uint8
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel context.CancelFunc
	logger log.FieldLogger
}

// converts our config struct for field to a telegraf field
func getTelegrafFields(incoming []Field) ([]telegrafPlugin.Field, error) {
	// initialize telegraf fields
	fields := make([]telegrafPlugin.Field, 0, len(incoming))

	// copy fields to table
	for i := range incoming {
		f := telegrafPlugin.Field{}
		if err := deepcopier.Copy(&incoming[i]).To(&f); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	return fields, nil
}

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})

	plugin := telegrafInputs.Inputs["snmp"]().(*telegrafPlugin.Snmp)

	// create the emitter
	em := baseemitter.NewEmitter(m.Output, m.logger)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	em.AddTag("plugin", strings.Replace(monitorType, "/", "-", -1))

	// create the accumulator
	ac := accumulator.NewAccumulator(em)

	// copy configurations to the plugin
	if err = deepcopier.Copy(conf).To(plugin); err != nil {
		m.logger.Error("unable to copy configurations to plugin")
		return err
	}

	// if a service is discovered that exposes snmp, take the host and port and add them to the agents list
	if conf.Host != "" {
		if plugin.Agents == nil {
			plugin.Agents = []string{fmt.Sprintf("%s:%d", conf.Host, conf.Port)}
		} else {
			plugin.Agents = append(plugin.Agents, fmt.Sprintf("%s:%d", conf.Host, conf.Port))
		}
	}

	// get top level telegraf fields
	plugin.Fields, err = getTelegrafFields(conf.Fields)
	if err != nil {
		return err
	}

	// initialize plugin.Tables
	plugin.Tables = make([]telegrafPlugin.Table, 0, len(conf.Tables))

	// copy tables
	for i := range conf.Tables {
		table := conf.Tables[i]
		t := telegrafPlugin.Table{}
		if err := deepcopier.Copy(&table).To(&t); err != nil {
			return err
		}

		// get telegraf fields
		t.Fields, err = getTelegrafFields(table.Fields)
		if err != nil {
			return err
		}

		plugin.Tables = append(plugin.Tables, t)
	}

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// gather metrics on the specified interval
	utils.RunOnInterval(ctx, func() {
		if err := plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return err
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
