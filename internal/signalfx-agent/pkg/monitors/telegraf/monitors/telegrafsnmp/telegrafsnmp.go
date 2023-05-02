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
	// Name of the field.  The OID will be used if no value is supplied.
	Name string `yaml:"name"`
	// The OID to fetch.
	Oid string `yaml:"oid"`
	// The sub-identifier to strip off when matching indexes to other fields.
	OidIndexSuffix string `yaml:"oidIndexSuffix"`
	// The index length after the table OID.  The index will be truncated after
	// this length in order to remove length index suffixes or non-fixed values.
	OidIndexLength int `yaml:"oidIndexLength"`
	// Whether to output the field as a tag.
	IsTag bool `yaml:"isTag"`
	// Controls the type conversion applied to the value: `"float(X)"`, `"float"`,
	// `"int"`, `"hwaddr"`, `"ipaddr"` or `""` (default).
	Conversion string `yaml:"conversion"`
}

// Table represents an SNMP table
type Table struct {
	// Metric name.  If not supplied the OID will be used.
	Name string `yaml:"name"`
	// Top level tags to inherit.
	InheritTags []string `yaml:"inheritTags"`
	// Add a tag for the table index for each row.
	IndexAsTag bool `yaml:"indexAsTag"`
	// Specifies the ags and values to look up.
	Fields []Field `yaml:"field"`
	// The OID to fetch.
	Oid string `yaml:"oid"`
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true"`
	// Host and port will be concatenated and appended to the list of SNMP agents to connect to.
	Host string `yaml:"host"`
	// Port and Host will be concatenated and appended to the list of SNMP agents to connect to.
	Port uint16 `yaml:"port"`
	// SNMP agent address and ports to query for information.  An example address is `0.0.0.0:5555`
	// If an address is supplied with out a port, the default port `161` will be used.
	Agents []string `yaml:"agents"`
	// The number of times to retry.
	Retries int `yaml:"retries"`
	// The SNMP protocol version to use (ie: `1`, `2`, `3`).
	Version uint8
	// The SNMP community to use.
	Community string `yaml:"community" default:"public"`
	// Maximum number of iterations for reqpeating variables
	MaxRepetitions uint8 `yaml:"maxRepetitions" default:"50"`
	// SNMP v3 context name to use with requests
	ContextName string `yaml:"contextName"`
	// Security level to use for SNMP v3 messages: `noAuthNoPriv` `authNoPriv`, `authPriv`.
	SecLevel string `yaml:"secLevel" default:"noAuthNoPriv"`
	// Name to used to authenticate with SNMP v3 requests.
	SecName string `yaml:"secName"`
	// Protocol to used to authenticate SNMP v3 requests: `"MD5"`, `"SHA"`, and `""` (default).
	AuthProtocol string `yaml:"authProtocol" default:""`
	// Password used to authenticate SNMP v3 requests.
	AuthPassword string `yaml:"authPassword" default:"" neverLog:"true"`
	// Protocol used for encrypted SNMP v3 messages: `DES`, `AES`, `""` (default).
	PrivProtocol string `yaml:"privProtocol" default:""`
	// Password used to encrypt SNMP v3 messages.
	PrivPassword string `yaml:"privPassword"`
	// The SNMP v3 engine ID.
	EngineID string `yaml:"engineID"`
	// The SNMP v3 engine boots.
	EngineBoots uint32 `yaml:"engineBoots"`
	// The SNMP v3 engine time.
	EngineTime uint32 `yaml:"engineTime"`
	// The top-level measurement name
	Name string `yaml:"name"`
	// The top-level SNMP fields
	Fields []Field `yaml:"fields"`
	// SNMP Tables
	Tables []Table `yaml:"tables"`
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
