// Copyright The OpenTelemetry Authors
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
	"fmt"
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

// NewConfig creates a new stdin input config with default values
func NewConfig() *Config {
	return &Config{
		InputConfig: helper.NewInputConfig(operatorType, operatorType),
		BaseConfig: BaseConfig{
			Multiline: helper.NewMultilineConfig(),
			Encoding:  helper.NewEncodingConfig(),
		},
	}
}

// Config is the configuration of a stdin input operator.
type Config struct {
	helper.InputConfig `mapstructure:",squash"`
	BaseConfig         `mapstructure:",squash"`
}

// BaseConfig is the detailed configuration of a tcp input operator.
type BaseConfig struct {
	Multiline          helper.MultilineConfig `mapstructure:"multiline,omitempty"`
	ScriptName         string                 `mapstructure:"script_name,omitempty"`
	Encoding           helper.EncodingConfig  `mapstructure:",squash,omitempty"`
	Source             string                 `mapstructure:"source"`
	SourceType         string                 `mapstructure:"sourcetype"`
	CollectionInterval string                 `mapstructure:"collection_interval"`
	MaxLogSize         helper.ByteSize        `mapstructure:"max_log_size,omitempty"`
	AddAttributes      bool                   `mapstructure:"add_attributes,omitempty"`
}

// Build will build a stdin input operator.
func (c *Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	if c.ScriptName == "" {
		return nil, fmt.Errorf("'exec_file' must be specified")
	}

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

	if c.MaxLogSize < minMaxLogSize {
		return nil, fmt.Errorf(
			"invalid value for parameter 'max_log_size', must be equal to or greater than %d bytes", minMaxLogSize,
		)
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
func (g *Input) Start(_ operator.Persister) error {

	ctx, cancelAll := context.WithCancel(context.Background())
	g.cancelAll = cancelAll

	ticker, err := createTicker(g.baseConfig.CollectionInterval)
	if err != nil {
		// TODO: move ticker verification to Start() so that we cannot error here.
		return err
	}

	go func() {
		for {
			_, cancelCycle := context.WithCancel(ctx)

			err := g.beginCycle(ctx)
			if err != nil {
				g.logger.Errorf("Error running script: %v", err)
			}

			select {
			case <-ctx.Done():
				cancelCycle()
				return
			case <-ticker.C:
				cancelCycle()
				g.wg.Wait()
			}
		}
	}()

	return nil
}

func (g *Input) beginCycle(ctx context.Context) error {

	stdOutReader, stdOutWriter := io.Pipe()
	commander, err := NewCommander(g.logger.Desugar(), g.baseConfig.ScriptName, stdOutWriter)
	if err != nil {
		return err
	}

	if err := commander.Start(ctx); err != nil {
		return err
	}

	g.wg.Add(2)

	readerCtx, cancelReader := context.WithCancel(ctx)

	go func() {
		defer g.wg.Done()
		select {
		case <-commander.Done():
			g.logger.Debug("Script finished", zap.String("exec_file", g.baseConfig.ScriptName))
			// Close the write pipe. This will result in subsequent read by scanner to return EOF and finish
			// the goroutine that processes the script output.
			err := stdOutWriter.Close()
			if err != nil {
				return
			}

		case <-ctx.Done():
			g.logger.Info("Script run too long. Stopping.", zap.String("exec_file", g.baseConfig.ScriptName))
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

	go g.readOutput(readerCtx, stdOutReader)

	return nil
}

func (g *Input) readOutput(ctx context.Context, r io.Reader) {
	defer g.wg.Done()

	buf := make([]byte, 0, g.MaxLogSize)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(buf, g.MaxLogSize)

	scanner.Split(g.splitFunc)

	for scanner.Scan() {
		decoded, err := g.encoding.Decode(scanner.Bytes())
		if err != nil {
			g.Errorw("Failed to decode data", zap.Error(err))
			continue
		}

		entry, err := g.NewEntry(string(decoded))
		if err != nil {
			g.Errorw("Failed to create entry", zap.Error(err))
			continue
		}

		if g.baseConfig.Source != "" {
			entry.AddAttribute("com.splunk.source", g.baseConfig.Source)
		}
		if g.baseConfig.SourceType != "" {
			entry.AddAttribute("com.splunk.sourcetype", g.baseConfig.SourceType)
		}

		g.Write(ctx, entry)
	}
	if err := scanner.Err(); err != nil {
		g.Errorw("Scanner error", zap.Error(err))
	}
}

// Stop will stop generating logs.
func (g *Input) Stop() error {
	g.cancelAll()
	g.wg.Wait()
	return nil
}
