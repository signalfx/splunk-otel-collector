// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scriptedinputsreceiver

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

func init() {
	operator.Register(operatorType, func() operator.Builder { return createDefaultConfig() })
}

const (
	// DefaultMaxLogSize is the max buffer sized used
	// if MaxLogSize is not set
	defaultMaxLogSize = 1024 * 1024
	operatorType      = "procpipe_input"
)

// stdoutOperator is an operator that reads input from stdout
type stdoutOperator struct {
	helper.InputOperator
	cfg           *Config
	logger        *zap.SugaredLogger
	cancelAll     context.CancelFunc
	splitFunc     bufio.SplitFunc
	decoder       *helper.Decoder
	baseConfig    adapter.BaseConfig
	wg            sync.WaitGroup
	scriptContent string
}

// Start will start generating log entries.
func (i *stdoutOperator) Start(_ operator.Persister) error {

	ctx, cancelAll := context.WithCancel(context.Background())
	i.cancelAll = cancelAll

	ticker := time.NewTicker(i.cfg.interval)

	go func() {
		for {
			_, cancelCycle := context.WithCancel(ctx)

			err := i.beginCycle(ctx)
			if err != nil {
				i.logger.Errorf("Error running script: %v", err)
			}

			select {
			case <-ctx.Done():
				cancelCycle()
				return
			case <-ticker.C:
				cancelCycle()
				i.wg.Wait()
			}
		}
	}()

	return nil
}

func (i *stdoutOperator) beginCycle(ctx context.Context) error {
	stdOutReader, stdOutWriter := io.Pipe()
	commander, err := newCommander(i.logger.Desugar(), i.cfg.ScriptName, i.scriptContent, stdOutWriter)
	if err != nil {
		return err
	}

	if err := commander.Start(ctx); err != nil {
		return err
	}

	i.wg.Add(2)

	readerCtx, cancelReader := context.WithCancel(ctx)

	go func() {
		defer i.wg.Done()
		select {
		case <-commander.Done():
			i.logger.Debug("Script finished", zap.String("script_name", i.cfg.ScriptName))
			// Close the write pipe. This will result in subsequent read by scanner to return EOF and finish
			// the goroutine that processes the script output.
			err := stdOutWriter.Close()
			if err != nil {
				return
			}

		case <-ctx.Done():
			i.logger.Info("Script run too long. Stopping.", zap.String("script_name", i.cfg.ScriptName))
			err := commander.Stop(context.Background())
			if err != nil {
				return
			}
			err2 := stdOutWriter.Close()
			if err2 != nil {
				return
			}
		}
		cancelReader()
	}()

	go i.readOutput(readerCtx, stdOutReader)

	return nil
}

func (i *stdoutOperator) readOutput(ctx context.Context, r io.Reader) {
	defer i.wg.Done()

	buf := make([]byte, 0, i.cfg.MaxLogSize)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(buf, int(i.cfg.MaxLogSize))

	scanner.Split(i.splitFunc)

	for scanner.Scan() {
		decoded, err := i.decoder.Decode(scanner.Bytes())
		if err != nil {
			i.Errorw("Failed to decode data", zap.Error(err))
			continue
		}

		entry, err := i.NewEntry(string(decoded))
		if err != nil {
			i.Errorw("Failed to create entry", zap.Error(err))
			continue
		}

		if i.cfg.Source != "" {
			entry.AddAttribute("com.splunk.source", i.cfg.Source)
		}
		if i.cfg.SourceType != "" {
			entry.AddAttribute("com.splunk.sourcetype", i.cfg.SourceType)
		}

		i.Write(ctx, entry)
	}
	if err := scanner.Err(); err != nil {
		i.Errorw("Scanner error", zap.Error(err))
	}
}

// Stop will stop generating logs.
func (i *stdoutOperator) Stop() error {
	i.cancelAll()
	i.wg.Wait()
	return nil
}
