// Copyright Copyright Splunk, Inc.
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

package procpipe

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

const (
	operatorType = "procpipe_input"

	// minMaxLogSize is the minimal size which can be used for buffering
	// TCP input
	minMaxLogSize = 64 * 1024

	// DefaultMaxLogSize is the max buffer sized used
	// if MaxLogSize is not set
	DefaultMaxLogSize = 1024 * 1024

	DefaultIntervalSeconds = 60
)

func init() {
	operator.Register(operatorType, func() operator.Builder { return NewConfig() })
}

// Build will build a stdin input operator.
func (c *Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	// If MaxLogSize not set, set sane default
	if c.MaxLogSize == 0 {
		c.MaxLogSize = DefaultMaxLogSize
	}

	if c.CollectionInterval == "" {
		c.CollectionInterval = strconv.Itoa(DefaultIntervalSeconds)
	}

	encoding, err := c.Encoding.Build()
	if err != nil {
		return nil, err
	}

	// Build multiline
	var splitFunc bufio.SplitFunc
	if c.Multiline.LineStartPattern == "" && c.Multiline.LineEndPattern == "" {
		splitFunc = helper.SplitNone(int(c.MaxLogSize))
	} else {
		splitFunc, err = c.Multiline.Build(encoding.Encoding, true, false, false, nil, int(c.MaxLogSize))
		if err != nil {
			return nil, err
		}
	}

	return &Input{
		baseConfig:    c.BaseConfig,
		InputOperator: inputOperator,
		logger:        logger,
		MaxLogSize:    int(c.MaxLogSize),
		encoding:      encoding,
		splitFunc:     splitFunc,
	}, nil
}

// Input is an operator that reads input from stdin
type Input struct {
	logger    *zap.SugaredLogger
	cancelAll context.CancelFunc
	splitFunc bufio.SplitFunc
	helper.InputOperator
	encoding   helper.Encoding
	baseConfig BaseConfig
	wg         sync.WaitGroup
	MaxLogSize int
}

func createTicker(intervalStr string) (*time.Ticker, error) {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, err
	}
	return time.NewTicker(interval), nil
}

// Start will start generating log entries.
func (i *Input) Start(_ operator.Persister) error {

	ctx, cancelAll := context.WithCancel(context.Background())
	i.cancelAll = cancelAll

	ticker, err := createTicker(i.baseConfig.CollectionInterval)
	if err != nil {
		// TODO: move ticker verification to Start() so that we cannot error here.
		return err
	}

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

func (i *Input) beginCycle(ctx context.Context) error {

	stdOutReader, stdOutWriter := io.Pipe()
	commander, err := NewCommander(i.logger.Desugar(), i.baseConfig.ScriptName, stdOutWriter)
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
			i.logger.Debug("Script finished", zap.String("script_name", i.baseConfig.ScriptName))
			// Close the write pipe. This will result in subsequent read by scanner to return EOF and finish
			// the goroutine that processes the script output.
			err := stdOutWriter.Close()
			if err != nil {
				return
			}

		case <-ctx.Done():
			i.logger.Info("Script run too long. Stopping.", zap.String("script_name", i.baseConfig.ScriptName))
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

func (i *Input) readOutput(ctx context.Context, r io.Reader) {
	defer i.wg.Done()

	buf := make([]byte, 0, i.MaxLogSize)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(buf, i.MaxLogSize)

	scanner.Split(i.splitFunc)

	for scanner.Scan() {
		decoded, err := i.encoding.Decode(scanner.Bytes())
		if err != nil {
			i.Errorw("Failed to decode data", zap.Error(err))
			continue
		}

		entry, err := i.NewEntry(string(decoded))
		if err != nil {
			i.Errorw("Failed to create entry", zap.Error(err))
			continue
		}

		if i.baseConfig.Source != "" {
			entry.AddAttribute("com.splunk.source", i.baseConfig.Source)
		}
		if i.baseConfig.SourceType != "" {
			entry.AddAttribute("com.splunk.sourcetype", i.baseConfig.SourceType)
		}

		i.Write(ctx, entry)
	}
	if err := scanner.Err(); err != nil {
		i.Errorw("Scanner error", zap.Error(err))
	}
}

// Stop will stop generating logs.
func (i *Input) Stop() error {
	i.cancelAll()
	i.wg.Wait()
	return nil
}
