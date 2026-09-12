// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package scriptedinput

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
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

// ScriptedInput is an operator that executes scripts and processes their output.
type ScriptedInput struct {
	logger   *zap.Logger
	doneChan chan struct{}
	command  *exec.Cmd
	cfg      Config
	helper.InputOperator
}

// Start starts the ScriptedInput.
func (si *ScriptedInput) Start(_ operator.Persister) error {
	if _, err := si.scheduleInput(si.cfg.BaseDir, si.cfg.Input); err != nil {
		return err
	}
	return nil
}

// Stop stops the ScriptedInput.
func (si *ScriptedInput) Stop() error {
	if si.command != nil {
		_ = si.command.Process.Signal(syscall.SIGTERM)
	}
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

	stopRead := make(chan struct{})
	go func() {
		for {
			select {
			case <-si.doneChan:
				return
			case <-stopRead:
				return
			default:
				b, ioErr := io.ReadAll(stdout)
				if len(b) == 0 {
					return
				}
				e := entry.New()
				e.Body = string(b)
				if err := si.Attribute(e); err != nil {
					si.logger.Error("Error setting attributes", zap.Error(err))
				}

				if err = si.Write(context.Background(), e); err != nil {
					si.logger.Error("Error consuming logs", zap.Error(err))
				}
				if ioErr != nil {
					return
				}
			}
		}
	}()

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
	si.command = cmd

	err = cmd.Wait()
	close(stopRead)

	return err
}
