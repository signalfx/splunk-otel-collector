package statsd

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type statsDListener struct {
	sync.Mutex
	ipAddr         string
	port           uint16
	tcp            bool
	udpConn        *net.UDPConn
	prefix         string
	converters     []*converter
	metricBuffer   []string
	shutdownCalled int32
	logger         *utils.ThrottledLogger
}

type statsDMetric struct {
	rawMetricName string
	metricName    string
	metricType    string
	value         float64
	dimensions    map[string]string
}

func (sl *statsDListener) Listen() error {
	if sl.tcp {
		return sl.listenTCP()
	}

	return sl.listenUDP()
}

func (sl *statsDListener) listenUDP() error {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(sl.ipAddr),
		Port: int(sl.port),
	})

	if err != nil {
		return err
	}

	sl.logger.Infof("SignalFx StatsD monitor: Listening on host & port %s:%s", conn.LocalAddr().Network(), conn.LocalAddr().String())

	sl.udpConn = conn
	return nil
}

func (sl *statsDListener) listenTCP() error {
	return nil
}

func (sl *statsDListener) FetchMetrics() []*statsDMetric {
	sl.Lock()
	rawMetrics := make([]string, len(sl.metricBuffer))

	copy(rawMetrics, sl.metricBuffer)
	sl.metricBuffer = nil
	sl.Unlock()

	parsed := sl.parseMetrics(rawMetrics)

	return parsed
}

func (sl *statsDListener) Read() {
	chData := make(chan []byte)

	if sl.tcp {
		go sl.readTCP(chData)
	} else {
		go sl.readUDP(chData)
	}

	for data := range chData {
		sl.Lock()
		sl.metricBuffer = append(sl.metricBuffer, strings.Split(string(data), "\n")...)
		sl.Unlock()
	}
}

func (sl *statsDListener) readTCP(chData chan []byte) {
}

func (sl *statsDListener) readUDP(chData chan []byte) {
	// UDP needs to receive data packet by packet. Max packet size is 65535 for now.
	buf := make([]byte, 65536)
	for {
		n, _, err := sl.udpConn.ReadFromUDP(buf)

		if err != nil {
			// Exit the loop if the connection is closed
			if atomic.LoadInt32(&sl.shutdownCalled) > 0 {
				break
			}

			sl.logger.WithError(err).Error("Failed reading UDP datagram.")
			continue
		}

		received := make([]byte, n)
		copy(received, buf[0:n])

		chData <- received
	}
}

func (sl *statsDListener) Close() {
	if !sl.tcp {
		atomic.StoreInt32(&sl.shutdownCalled, 1)
		sl.udpConn.Close()
	}
}

func (sl *statsDListener) parseMetrics(raw []string) []*statsDMetric {
	var metrics []*statsDMetric

	for _, m := range raw {
		if m == "" {
			continue
		}
		m, dims := parseDogstatsdTags(m, sl.logger)
		colonIdx := strings.Index(m, ":")
		pipeIdx := strings.Index(m, "|")
		if pipeIdx >= len(m)-1 || pipeIdx < 0 || colonIdx-1 > len(m) || colonIdx < 0 {
			sl.logger.Warnf("Invalid StatsD metric string : %s", m)
			continue
		}
		secondPipeIdx := pipeIdx + strings.Index(m[pipeIdx+1:], "|")

		rawMetricName := m[0:colonIdx]

		var metricType string
		if secondPipeIdx > pipeIdx {
			metricType = m[pipeIdx+1 : secondPipeIdx]
		} else {
			metricType = m[pipeIdx+1:]
		}

		var metricName string
		if sl.prefix != "" {
			metricName = strings.TrimPrefix(rawMetricName, sl.prefix+".")
		} else {
			metricName = rawMetricName
		}

		if sl.converters != nil {
			metricName, dims = convertMetric(metricName, sl.converters, dims)
		}

		strValue := m[colonIdx+1 : pipeIdx]
		value, err := strconv.ParseFloat(strValue, 64)

		if err == nil {
			metrics = append(metrics, &statsDMetric{
				rawMetricName: rawMetricName,
				metricName:    metricName,
				metricType:    metricType,
				value:         value,
				dimensions:    dims,
			})
		} else {
			sl.logger.WithError(err).Errorf("Failed parsing metric value %s", strValue)
		}
	}

	return metrics
}
