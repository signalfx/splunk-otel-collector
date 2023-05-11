package processlist

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/event"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

const version = "0.0.30"

// EVENT(objects.top-info): Process list event.

var zlibCompressor = zlib.NewWriter(&bytes.Buffer{})
var now = time.Now

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"true" acceptsEndpoints:"false"`
}

func init() {
	monitors.Register(&monitorMetadata, func() interface{} {
		return &Monitor{}
	}, &Config{})
}

// compresses the given byte array
func compressBytes(in []byte) (buf bytes.Buffer, err error) {
	zlibCompressor.Reset(&buf)
	_, err = zlibCompressor.Write(in)
	_ = zlibCompressor.Close()
	return
}

// Monitor for Utilization
type Monitor struct {
	Output        types.Output
	cancel        func()
	lastCPUCounts map[procKey]time.Duration
	nextPurge     time.Time
	logger        log.FieldLogger
}

// TopProcess is a platform-independent way of representing a process to be
// reported to SignalFx
type TopProcess struct {
	ProcessID           int
	CreatedTime         time.Time
	Username            string
	Priority            int
	Nice                *int
	VirtualMemoryBytes  uint64
	WorkingSetSizeBytes uint64
	SharedMemBytes      uint64
	Status              string
	MemPercent          float64
	TotalCPUTime        time.Duration
	Command             string
}

type procKey string

// Key used to uniquely identify a process, even if pid is reused.
func (tp *TopProcess) key() procKey {
	return procKey(fmt.Sprintf("%d|%s", tp.ProcessID, tp.Command))
}

// Configure configures the monitor and starts collecting on the configured interval
func (m *Monitor) Configure(conf *Config) error {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	// create contexts for managing the plugin loop
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	interval := time.Duration(conf.IntervalSeconds) * time.Second

	m.lastCPUCounts = make(map[procKey]time.Duration)

	osCache := initOSCache()
	m.nextPurge = now().Add(3 * time.Minute)

	utils.RunOnInterval(
		ctx,
		func() {
			// get the process list
			procs, err := ProcessList(conf, osCache, m.logger)
			if err != nil {
				m.logger.WithError(err).Error("Couldn't get process list")
				return
			}

			message, err := m.encodeEventMessage(procs, interval)
			if err != nil {
				m.logger.WithError(err).Error("Failed to encode process list")
			}

			m.Output.SendEvent(
				&event.Event{
					EventType:  "objects.top-info",
					Category:   event.AGENT,
					Dimensions: map[string]string{},
					Properties: map[string]interface{}{
						"message": message,
					},
					Timestamp: time.Now(),
				},
			)

			curTime := now()
			if curTime.After(m.nextPurge) {
				m.purgeCPUCache(procs)
				m.nextPurge = curTime.Add(3 * time.Minute)
			}
		},
		interval,
	)
	return nil
}

func (m *Monitor) encodeEventMessage(procs []*TopProcess, sampleInterval time.Duration) (string, error) {
	if len(procs) == 0 {
		return "", errors.New("no processes to encode")
	}

	procsEncoded := []byte{'{'}
	for i := range procs {
		procsEncoded = append(procsEncoded, []byte(m.encodeProcess(procs[i], sampleInterval)+",")...)
	}
	procsEncoded[len(procsEncoded)-1] = '}'

	// escape and compress the process list
	escapedBytes := bytes.Replace(procsEncoded, []byte{byte('\\')}, []byte{byte('\\'), byte('\\')}, -1)
	compressedBytes, err := compressBytes(escapedBytes)
	if err != nil {
		return "", fmt.Errorf("couldn't compress process list: %v", err)
	}

	return fmt.Sprintf(
		"{\"t\":\"%s\",\"v\":\"%s\"}",
		base64.StdEncoding.EncodeToString(compressedBytes.Bytes()), version), nil
}

func (m *Monitor) encodeProcess(proc *TopProcess, sampleInterval time.Duration) string {
	key := proc.key()
	lastSampleInterval := sampleInterval
	lastCPUCount, ok := m.lastCPUCounts[key]
	if !ok {
		lastSampleInterval = time.Since(proc.CreatedTime)
	}
	m.lastCPUCounts[key] = proc.TotalCPUTime

	cpuPercent := float64(proc.TotalCPUTime-lastCPUCount) * 100.0 / float64(lastSampleInterval)

	var nice string
	if proc.Nice == nil {
		nice = "unknown"
	} else {
		nice = strconv.Itoa(*proc.Nice)
	}

	return fmt.Sprintf(`"%d":["%s",%d,"%s",%d,%d,%d,"%s",%.2f,%.2f,"%s","%s"]`,
		proc.ProcessID,
		strings.ReplaceAll(proc.Username, `"`, "'"),
		proc.Priority,
		nice,
		proc.VirtualMemoryBytes/1024,
		proc.WorkingSetSizeBytes/1024,
		proc.SharedMemBytes/1024,
		proc.Status,
		cpuPercent,
		proc.MemPercent,
		toTime(proc.TotalCPUTime.Seconds()),
		strings.ReplaceAll(proc.Command, `"`, `'`),
	)
}

func (m *Monitor) purgeCPUCache(lastProcs []*TopProcess) {
	lastKeys := make(map[procKey]struct{}, len(lastProcs))
	for i := range lastProcs {
		lastKeys[lastProcs[i].key()] = struct{}{}
	}

	for k := range m.lastCPUCounts {
		if _, ok := lastKeys[k]; !ok {
			delete(m.lastCPUCounts, k)
		}
	}
}

// toTime returns the given seconds as a formatted string "min:sec.dec"
func toTime(secs float64) string {
	minutes := int(secs) / 60
	seconds := math.Mod(secs, 60.0)
	dec := math.Mod(seconds, 1.0) * 100
	return fmt.Sprintf("%02d:%02.f.%02.f", minutes, seconds, dec)
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
