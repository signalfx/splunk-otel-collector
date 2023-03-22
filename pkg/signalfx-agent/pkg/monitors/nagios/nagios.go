package nagios

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/go-test/deep"
	"github.com/kballard/go-shellquote"
	"github.com/patrickmn/go-cache"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"false"`
	// The command to exec with any arguments like:
	// `"LC_ALL=\"en_US.utf8\" /usr/lib/nagios/plugins/check_ntp_time -H pool.ntp.typhon.net -w 0.5 -c 1"`
	Command string `yaml:"command" validate:"required"`
	// Corresponds to the nagios `service` column and allows to aggregate all
	// instances of the same service (when calling the same check script with
	// different arguments)
	Service string `yaml:"service" validate:"required"`
	// The max execution time allowed in seconds before sending SIGKILL
	Timeout int `yaml:"timeout" default:"9"`
	// If `false` and change is detected on `stdout` compared to the last
	// event it will send a new one
	IgnoreStdOut bool `yaml:"ignoreStdOut" default:"false"`
	// If `false` and change is detected on `stderr` compared to the last
	// event it will send a new one
	IgnoreStdErr bool `yaml:"ignoreStdErr" default:"false"`
}

// Monitor that collect metrics
type Monitor struct {
	Output types.FilteringOutput
	cancel func()
	logger logrus.FieldLogger
}

const (
	unknown          = 3
	cacheKey         = "lastRun"
	propertiesLength = 256
)

var (
	// ErrorTimeout is returned when command is killed after timeout duration
	ErrorTimeout = errors.New("command killed after timeout")
)

// Configure and kick off internal metric collection
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	// Define global dimensions used for both datapoint and event
	dimensions := map[string]string{
		"plugin":  "nagios",
		"command": conf.Command,
		"service": conf.Service,
	}
	// Enforce interval greater than command timeout
	if conf.IntervalSeconds < conf.Timeout {
		return fmt.Errorf("configured timeout must be lower than intervalSeconds")
	}
	// Init cache used to avoid sending duplicate event for each interval
	c := cache.New(1*time.Minute, 1*time.Hour)

	// Start the metric gathering process here
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	utils.RunOnInterval(ctx, func() {
		// Run command
		stdout, stderr, err := runCommand(conf.Command, conf.Timeout)
		state, err := getExitCode(err, stdout)

		// Send datapoint with only state
		m.Output.SendDatapoints([]*datapoint.Datapoint{
			datapoint.New(
				nagiosState,
				dimensions,
				datapoint.NewIntValue(int64(state)),
				datapoint.Gauge,
				time.Time{}),
		}...)

		properties := makeProperties(state, err, stdout, stderr)
		// Some scripts could produce different output (and stderr) for
		// each interval "normally", so we do not want to compare them
		diffProperties := filterProperties(properties, conf.IgnoreStdOut, conf.IgnoreStdErr)

		// Compare with previous event if it exists
		sendEvent := true
		if x, found := c.Get(cacheKey); found {
			lastProperties := x.(map[string]interface{})
			if diff := deep.Equal(diffProperties, lastProperties); diff == nil {
				m.logger.Debug("the same event has already been sent, do not send again")
				sendEvent = false
			}
		}

		// Do not send duplicate event
		if sendEvent {
			// Send event with command context
			m.Output.SendEvent(
				event.NewWithProperties(
					nagiosState,
					event.AGENT,
					dimensions,
					properties,
					time.Time{}),
			)
			// update event properties in cache
			c.Set(cacheKey, diffProperties, cache.NoExpiration)
		}

	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown the monitor
func (m *Monitor) Shutdown() {
	// Stop any long-running go routines here
	if m.cancel != nil {
		m.cancel()
	}
}

func runCommand(command string, timeout int) (stdout []byte, stderr []byte, err error) {
	var cmdOut, cmdErr bytes.Buffer

	// Parse command string with args
	splitCmd, err := shellquote.Split(command)
	if err != nil || len(splitCmd) == 0 {
		return nil, nil, fmt.Errorf("exec: unable to parse command, %s", err)
	}
	// Prepare command exec
	cmd := exec.Command(splitCmd[0], splitCmd[1:]...)
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	// Start command
	if err = cmd.Start(); err != nil {
		return
	}

	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	// Start a timer
	killTimeout := time.After(time.Duration(timeout) * time.Second)

	// The select statement allows us to execute based on which channel
	// we get a message from first.
	select {
	case <-killTimeout:
		// Timeout happened first, kill the process and print a message.
		err = cmd.Process.Kill()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, ErrorTimeout
	case err := <-done:
		// Command completed before timeout. Print output and error if it exists.
		return cmdOut.Bytes(), cmdErr.Bytes(), err

	}
}

func getExitCode(err error, stdout []byte) (int, error) {
	// See https://nagios-plugins.org/doc/guidelines.html#AEN78
	// Some scripts could not respect the nagios convention and
	// returns a code == 0 even if it show an error message.
	// To mitigate this risk we override an anormal state when
	// information allow to think the exit code is not relevant
	if err == nil {
		// We only need the first 3 bytes to determine the status
		// from the string output (over the exit code)
		status := strings.ToLower(string(stdout[:4]))
		switch status {
		case "crit":
			return 2, nil
		case "warn":
			return 1, nil
		case "unkn":
			return unknown, nil
		}
		return 0, nil
	}
	// If there is an error but we cannot get the exit code so
	// we cannot really consider it is a "critical" or a "warning"
	// and we will use the "unknown" state
	ee, ok := err.(*exec.ExitError)
	if !ok {
		// If killed the error does not make sens
		if err == ErrorTimeout {
			return unknown, err
		}
		return unknown, fmt.Errorf("command ended unexpectedly %s", err)
	}

	ws, ok := ee.Sys().(syscall.WaitStatus)
	if !ok {
		return 3, fmt.Errorf("cannot get exit code")
	}

	// An error with the exit code so we can be sure the command has been
	// executed to its end even if the exitcode still can be an error in
	// the script (not the nagios state). In any case we return the exit
	// code as state and the user could find more information in the event
	return ws.ExitStatus(), nil
}

func makeProperties(state int, err error, stdout []byte, stderr []byte) map[string]interface{} {
	properties := make(map[string]interface{})
	if len(stdout) > 0 {
		properties["stdout"] = formatStd(stdout)
	}
	if len(stderr) > 0 {
		properties["stderr"] = formatStd(stderr)
	}
	if err != nil {
		properties["exit_reason"] = err.Error()
	} else {
		properties["exit_code"] = state
	}

	return properties
}

func formatStd(std []byte) string {
	rendered := strings.TrimLeft(strings.TrimSuffix(string(std), "\n"), "\n")
	if len(rendered) > propertiesLength {
		rendered = rendered[:propertiesLength]
	}
	return rendered
}

func filterProperties(properties map[string]interface{}, ignoreStdOut bool, ignoreStdErr bool) map[string]interface{} {
	filteredProperties := make(map[string]interface{})
	for k, v := range properties {
		if k == "stdout" && ignoreStdOut {
			continue
		}
		if k == "stderr" && ignoreStdErr {
			continue
		}
		filteredProperties[k] = v
	}

	return filteredProperties
}
