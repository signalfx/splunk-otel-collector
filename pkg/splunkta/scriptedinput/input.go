// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package scriptedinput

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/script"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
)

type ScriptedInput struct {
	logger   *zap.Logger
	doneChan chan struct{}
	command  *exec.Cmd
	cfg      Config
	helper.InputOperator
	mu sync.Mutex
}

func (si *ScriptedInput) Start(_ operator.Persister) error {
	if _, err := si.scheduleInput(si.cfg.BaseDir, si.cfg.Input); err != nil {
		return err
	}
	return nil
}

func (si *ScriptedInput) Stop() error {
	si.mu.Lock()
	if si.command != nil {
		_ = si.command.Process.Signal(syscall.SIGTERM)
	}
	si.mu.Unlock()
	close(si.doneChan)

	return nil
}

func (si *ScriptedInput) scheduleInput(baseDir string, input conf.Input) (bool, error) {
	parsed, err := stanza.ParseName(input.Configuration.Stanza.Name)
	if err != nil {
		return false, err
	}
	switch parsed.Kind {
	case "script":
		return si.scheduleScriptedInput(baseDir, input)
	case "":
		return si.scheduleScriptedInput(baseDir, input)
	default:
		return false, fmt.Errorf("unknown scheme %q", parsed.Kind)
	}
}

func (si *ScriptedInput) scheduleScriptedInput(baseDir string, input conf.Input) (bool, error) {
	intervalS := 3600.0
	for _, p := range input.Configuration.Stanza.Params {
		if p.Name == "interval" {
			var err error
			intervalS, err = strconv.ParseFloat(p.Value, 64)
			if err != nil {
				// TODO: cron schedule support not yet implemented
				return false, err
			}
		}
		if p.Name == "disabled" && p.Value == "1" {
			return false, nil
		}
	}
	if intervalS == -1 {
		return false, nil
	}
	if intervalS == 0 {
		go func() {
			for {
				select {
				case <-si.doneChan:
					return
				default:
					si.execute(baseDir, input)
				}
			}
		}()
	} else {
		interval := time.Duration(intervalS * float64(time.Second))
		go func() {
			si.execute(baseDir, input)

			for {
				select {
				case <-time.After(interval):
					si.execute(baseDir, input)
				case <-si.doneChan:
					return
				}
			}
		}()
	}
	return true, nil
}

func (si *ScriptedInput) execute(baseDir string, input conf.Input) {
	if err := si._execute(baseDir, input); err != nil {
		si.logger.Error("Error executing input", zap.String("input", input.Configuration.Stanza.Name), zap.String("error", err.Error()))
	}
}

func (si *ScriptedInput) _execute(baseDir string, input conf.Input) error {
	command, err := script.DetermineCommandName(baseDir, input)
	if err != nil {
		return err
	}
	cmd := exec.Command(command)
	var stdin io.WriteCloser
	var stdout io.ReadCloser
	if stdin, err = cmd.StdinPipe(); err != nil {
		return err
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return err
	}

	var inputXML []byte
	if inputXML, err = input.ToXML(); err != nil {
		return err
	}
	if _, err = stdin.Write(inputXML); err != nil {
		return err
	}
	if err = stdin.Close(); err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}
	si.mu.Lock()
	si.command = cmd
	si.mu.Unlock()

	// Read stdout to EOF before Wait: exec.Cmd.StdoutPipe closes the pipe once
	// the process exits, so reads must complete first. Reading synchronously
	// here (rather than in a separate goroutine racing cmd.Wait) avoids both
	// the data race and dropped output when the process exits quickly.
	b, readErr := io.ReadAll(stdout)
	if len(b) > 0 {
		e := entry.New()
		e.Body = string(b)
		if attrErr := si.Attribute(e); attrErr != nil {
			si.logger.Error("Error setting attributes", zap.Error(attrErr))
		}
		if writeErr := si.Write(context.Background(), e); writeErr != nil {
			si.logger.Error("Error consuming logs", zap.Error(writeErr))
		}
	}

	waitErr := cmd.Wait()
	if readErr != nil {
		return readErr
	}
	return waitErr
}
