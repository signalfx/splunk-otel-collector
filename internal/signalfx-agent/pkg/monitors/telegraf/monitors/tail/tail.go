package tail

import (
	"context"
	"strings"
	"time"

	telegrafInputs "github.com/influxdata/telegraf/plugins/inputs"
	telegrafPlugin "github.com/influxdata/telegraf/plugins/inputs/tail"
	telegrafParsers "github.com/influxdata/telegraf/plugins/parsers"
	log "github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/accumulator"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/emitter/baseemitter"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/common/parser"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

var logger = log.WithFields(log.Fields{"monitorType": monitorType})

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	parser               telegrafParsers.Parser
	TelegrafParser       *parser.Config `yaml:"telegrafParser"`
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"false"`
	WatchMethod          string   `yaml:"watchMethod" default:"poll"`
	Files                []string `yaml:"files" validate:"required"`
	Pipe                 bool     `yaml:"pipe" default:"false"`
	FromBeginning        bool     `yaml:"fromBeginning" default:"false"`
}

// Monitor for Utilization
type Monitor struct {
	Output types.Output
	cancel context.CancelFunc
	plugin *telegrafPlugin.Tail
	logger log.FieldLogger
}

// fetch the factory function used to generate the plugin
var factory = telegrafInputs.Inputs["tail"]

// Configure the monitor and kick off metric syncing
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = logger.WithField("monitorID", conf.MonitorID)
	m.plugin = factory().(*telegrafPlugin.Tail)

	// use the default config
	if conf.TelegrafParser == nil {
		m.logger.Debug("defaulting to influx parser because no parser was specified")
		conf.TelegrafParser = &parser.Config{DataFormat: "influx"}
	}

	// test the parser configurations to make sure they're valid
	if conf.parser, err = conf.TelegrafParser.GetTelegrafParser(); err != nil {
		return err
	}

	// copy configurations to the plugin
	if err = deepcopier.Copy(conf).To(m.plugin); err != nil {
		m.logger.Error("unable to copy configurations to plugin")
		return err
	}

	// set the parser on the plugin
	m.plugin.SetParserFunc(conf.TelegrafParser.GetTelegrafParser)

	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	// craete the emitter
	em := baseemitter.NewEmitter(m.Output, m.logger)

	// Hard code the plugin name because the emitter will parse out the
	// configured measurement name as plugin and that is confusing.
	em.AddTag("plugin", strings.ReplaceAll(monitorType, "/", "-"))

	// create the accumulator
	ac := accumulator.NewAccumulator(em)

	// start the tail plugin
	if err = m.plugin.Start(ac); err != nil {
		return err
	}

	// look for new files to tail on the defined interval
	utils.RunOnInterval(ctx, func() {
		if err := m.plugin.Gather(ac); err != nil {
			m.logger.WithError(err).Errorf("an error occurred while gathering metrics")
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		// stop the collection interval
		m.cancel()
	}
	if m.plugin != nil {
		// stop the telegraf plugin
		m.plugin.Stop()
	}
}
