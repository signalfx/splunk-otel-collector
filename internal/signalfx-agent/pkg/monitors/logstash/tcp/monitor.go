package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils/timeutil"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

type event map[string]interface{}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" acceptsEndpoints:"true" singleInstance:"false"`
	// If `mode: server`, the local IP address to listen on.  If `mode:
	// client`, the Logstash host/ip to connect to.
	Host string `yaml:"host" validate:"required"`
	// If `mode: server`, the local port to listen on.  If `mode: client`, the
	// port of the Logstash TCP output plugin.  If port is `0`, a random
	// listening port is assigned by the kernel.
	Port uint16 `yaml:"port"`

	// Whether to act as a `server` or `client`.  The corresponding setting in
	// the Logtash `tcp` output plugin should be set to the opposite of this.
	Mode string `yaml:"mode" default:"client" validate:"oneof=server client"`

	DesiredTimerFields []string `yaml:"desiredTimerFields" default:"[\"mean\",\"max\",\"p99\",\"count\"]"`
	// How long to wait before reconnecting if the TCP connection cannot be
	// made or after it gets broken.
	ReconnectDelay timeutil.Duration `yaml:"reconnectDelay" default:"5s"`
	// If true, events received from Logstash will be dumped to the agent's
	// stdout in deserialized form
	DebugEvents bool `yaml:"debugEvents"`
}

func (c *Config) DesiredTimerFieldSet() map[string]bool {
	return utils.StringSliceToMap(c.DesiredTimerFields)
}

// Monitor that accepts and forwards trace spans
type Monitor struct {
	Output types.Output
	conf   *Config
	ctx    context.Context
	cancel context.CancelFunc
	logger *utils.ThrottledLogger
}

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = utils.NewThrottledLogger(log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID}), 30*time.Second)

	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.conf = conf

	if conf.Mode == "server" {
		go m.keepReadingAsServer(conf.Host, conf.Port)
	} else {
		go m.keepReadingFromServer(conf.Host, conf.Port)
	}

	return nil
}

func (m *Monitor) keepReadingAsServer(host string, port uint16) {
OUTER:
	for {
		if m.ctx.Err() != nil {
			return
		}

		listener, err := net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.ParseIP(host),
			Port: int(port),
		})
		if err != nil {
			m.logger.WithFields(log.Fields{
				"err":  err,
				"host": host,
				"port": port,
			}).Error("Could not listen for Logstash events")
			time.Sleep(m.conf.ReconnectDelay.AsDuration())
			continue
		}

		m.logger.Infof("Listening for Logstash events on %s", listener.Addr().String())

		for {
			conn, err := listener.Accept()
			if err != nil {
				m.logger.WithError(err).Error("Could not accept Logstash connections")
				listener.Close()
				time.Sleep(m.conf.ReconnectDelay.AsDuration())
				continue OUTER
			}

			go func() {
				if err := m.handleConnection(conn); err != nil {
					if m.ctx.Err() == context.Canceled {
						return
					}
					m.logger.WithError(err).Info("Logstash inbound connection terminated")
				}
			}()
		}
	}
}

func (m *Monitor) keepReadingFromServer(host string, port uint16) {
	for {
		if m.ctx.Err() != nil {
			return
		}

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			m.logger.WithError(err).Error("Logstash TCP output connection failed")
			time.Sleep(m.conf.ReconnectDelay.AsDuration())
			continue
		}

		err = m.handleConnection(conn)
		if err != nil {
			if m.ctx.Err() == context.Canceled {
				return
			}
			m.logger.WithError(err).Error("Logstash receive failed")
			time.Sleep(m.conf.ReconnectDelay.AsDuration())
			continue
		}
	}
}

func (m *Monitor) handleConnection(conn net.Conn) error {
	go func() {
		<-m.ctx.Done()
		m.logger.Debug("Closing connection")
		conn.Close()
	}()

	var err error
	decoder := json.NewDecoder(bufio.NewReader(conn))
	for decoder.More() {
		var ev event
		err = decoder.Decode(&ev)
		if err != nil {
			err = fmt.Errorf("failed to parse logstash event: %v", err)
			break
		}

		dps, err := m.convertEventToDatapoints(ev)
		if err != nil {
			m.logger.WithError(err).Errorf("Failed to convert event to datapoints: %v", ev)
			continue
		}

		m.Output.SendDatapoints(dps...)
	}

	return err
}

func (m *Monitor) convertEventToDatapoints(ev event) ([]*datapoint.Datapoint, error) {
	if m.conf.DebugEvents {
		spew.Dump(ev)
	}

	timestamp := time.Time{}
	timestampStr, ok := ev["@timestamp"].(string)
	delete(ev, "@timestamp")
	if ok {
		var err error
		timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			m.logger.WithField("timestamp", timestampStr).Warn("Could not parse timestamp with RFC3339 format")
		}
	}

	var dps []*datapoint.Datapoint
	commonDims := map[string]string{}

	for k, v := range ev {
		metricMap, ok := v.(map[string]interface{})
		if !ok {
			commonDims[strings.ReplaceAll(k, "@", "")] = fmt.Sprintf("%v", v)
			continue
		}

		// Both timers and meters have the count field.
		_, ok = metricMap["count"].(float64)
		if !ok {
			m.logger.WithField("event", ev).Warnf("Saw event map without a 'count' field, is the Logstash event coming from the metrics filter?")
			continue
		}

		// The mean field indicates a timer value
		_, ok = metricMap["mean"].(float64)
		if ok {
			dps = append(dps, m.parseMapAsTimer(k, metricMap, m.conf.DesiredTimerFieldSet())...)
		} else {
			dps = append(dps, parseMapAsMeter(k, metricMap))
		}
	}

	if len(dps) == 0 {
		return nil, fmt.Errorf("no timer/meter metrics found in Logstash event (are you using metrics filter?)")
	}

	for i := range dps {
		dps[i].Dimensions = utils.MergeStringMaps(commonDims, dps[i].Dimensions)
		dps[i].Timestamp = timestamp
	}
	return dps, nil
}

func parseMapAsMeter(key string, metricMap map[string]interface{}) *datapoint.Datapoint {
	val, _ := datapoint.CastFloatValue(metricMap["count"].(float64))
	return datapoint.New(key+".count", nil, val, datapoint.Counter, time.Time{})
}

func (m *Monitor) parseMapAsTimer(key string, metricMap map[string]interface{}, desiredFields map[string]bool) []*datapoint.Datapoint {
	dps := make([]*datapoint.Datapoint, 0, len(desiredFields))
	for fieldName := range desiredFields {
		floatVal, ok := metricMap[fieldName].(float64)
		if !ok {
			m.logger.WithField("metric", metricMap).WithField("field", fieldName).Warn("Could not find desired field in map")
		}
		val, _ := datapoint.CastFloatValue(floatVal)
		typ := datapoint.Gauge
		if fieldName == "count" {
			typ = datapoint.Counter
		}
		dps = append(dps, datapoint.New(key+"."+fieldName, nil, val, typ, time.Time{}))
	}
	return dps
}

func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
