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

// termWaitDelay bounds how long a script has to exit after receiving SIGTERM
// on shutdown before it is force-killed, so Stop cannot hang indefinitely.
const termWaitDelay = 10 * time.Second

// ScriptedInput is an operator that executes scripts and processes their output.
type ScriptedInput struct {
	logger *zap.Logger
	cancel context.CancelFunc
	cfg    Config
	helper.InputOperator
	wg sync.WaitGroup
}

// Start starts the ScriptedInput.
func (si *ScriptedInput) Start(_ operator.Persister) error {
	ctx, cancel := context.WithCancel(context.Background())
	si.cancel = cancel
	if _, err := si.scheduleInput(ctx, si.cfg.BaseDir, si.cfg.Input); err != nil {
		cancel()
		return err
	}
	return nil
}

// Stop stops the ScriptedInput. It cancels any running script and waits for the
// scheduler goroutines to return.
func (si *ScriptedInput) Stop() error {
	if si.cancel != nil {
		si.cancel()
	}
	si.wg.Wait()
	return nil
}

func (si *ScriptedInput) scheduleInput(ctx context.Context, baseDir string, input conf.Input) (bool, error) {
	parsed, err := stanza.ParseName(input.Configuration.Stanza.Name)
	if err != nil {
		return false, err
	}
	switch parsed.Kind {
	case "script":
		return si.scheduleScriptedInput(ctx, baseDir, input)
	case "":
		return si.scheduleScriptedInput(ctx, baseDir, input)
	default:
		return false, fmt.Errorf("unknown scheme %q", parsed.Kind)
	}
}

func (si *ScriptedInput) scheduleScriptedInput(ctx context.Context, baseDir string, input conf.Input) (bool, error) {
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
	si.wg.Add(1)
	if intervalS == 0 {
		go func() {
			defer si.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					si.execute(ctx, baseDir, input)
				}
			}
		}()
	} else {
		interval := time.Duration(intervalS * float64(time.Second))
		go func() {
			defer si.wg.Done()
			si.execute(ctx, baseDir, input)

			for {
				select {
				case <-time.After(interval):
					si.execute(ctx, baseDir, input)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	return true, nil
}

func (si *ScriptedInput) execute(ctx context.Context, baseDir string, input conf.Input) {
	if err := si._execute(ctx, baseDir, input); err != nil {
		si.logger.Error("Error executing input", zap.String("input", input.Configuration.Stanza.Name), zap.String("error", err.Error()))
	}
}

func (si *ScriptedInput) _execute(ctx context.Context, baseDir string, input conf.Input) error {
	command, err := script.DetermineCommandName(baseDir, input)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, command)
	// Preserve SIGTERM-then-kill semantics: canceling ctx signals the process
	// with SIGTERM, and WaitDelay force-kills it if it does not exit in time.
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = termWaitDelay
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
			case <-ctx.Done():
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
				if attrErr := si.Attribute(e); attrErr != nil {
					si.logger.Error("Error setting attributes", zap.Error(attrErr))
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

	err = cmd.Wait()
	close(stopRead)

	return err
}
