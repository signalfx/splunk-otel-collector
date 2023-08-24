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
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

const (
	defaultCollectionInterval = "60s"
	// minMaxLogSize is the minimal size which can be used for buffering TCP input
	minMaxLogSize = 64 * 1024
)

var availableScripts = func() []string {
	var s []string
	for _, sn := range scripts {
		s = append(s, sn)
	}
	sort.Strings(s)
	return s
}()

type Config struct {
	helper.InputConfig `mapstructure:",squash"`
	Multiline          helper.MultilineConfig `mapstructure:"multiline,omitempty"`
	ScriptName         string                 `mapstructure:"script_name,omitempty"`
	Encoding           helper.EncodingConfig  `mapstructure:",squash,omitempty"`
	Source             string                 `mapstructure:"source"`
	SourceType         string                 `mapstructure:"sourcetype"`
	CollectionInterval string                 `mapstructure:"collection_interval"`
	MaxLogSize         helper.ByteSize        `mapstructure:"max_log_size,omitempty"`
	AddAttributes      bool                   `mapstructure:"add_attributes,omitempty"`
	interval           time.Duration
}

func createDefaultConfig() *Config {
	return &Config{
		InputConfig:        helper.NewInputConfig(typeStr, typeStr),
		Multiline:          helper.NewMultilineConfig(),
		Encoding:           helper.NewEncodingConfig(),
		CollectionInterval: defaultCollectionInterval,
		MaxLogSize:         defaultMaxLogSize,
	}
}

func (c *Config) Validate() error {
	if c.ScriptName == "" {
		return errors.New("'script_name' must be specified")
	}

	_, ok := scripts[c.ScriptName]
	if !ok {
		return fmt.Errorf("unsupported 'script_name' %q. must be one of ", c.ScriptName)
	}

	if c.MaxLogSize != 0 && c.MaxLogSize < minMaxLogSize {
		return fmt.Errorf("invalid value for parameter 'max_log_size', must be equal to or greater than %d bytes", minMaxLogSize)
	}

	if c.MaxLogSize > math.MaxInt {
		return fmt.Errorf("invalid value for parameter 'max_log_size', must be less than or equal to %d bytes", math.MaxInt)
	}

	var err error
	c.interval, err = time.ParseDuration(c.CollectionInterval)
	if err != nil {
		return fmt.Errorf("invalid 'collection_interval': %w", err)
	}

	return nil
}

// Build will build a stdoutOperator.
func (c *Config) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	inputOperator, err := c.InputConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	enc, err := helper.LookupEncoding(c.Encoding.Encoding)
	if err != nil {
		return nil, err
	}

	// Build multiline
	var splitFunc bufio.SplitFunc
	if c.Multiline.LineStartPattern == "" && c.Multiline.LineEndPattern == "" {
		splitFunc = helper.SplitNone(int(c.MaxLogSize))
	} else {
		splitFunc, err = c.Multiline.Build(enc, true, false, false, nil, int(c.MaxLogSize))
		if err != nil {
			return nil, err
		}
	}

	scriptContent, ok := scripts[c.ScriptName]
	if !ok {
		// should have already been detected
		return nil, fmt.Errorf("missing script %q", c.ScriptName)
	}

	return &stdoutOperator{
		cfg:           c,
		InputOperator: inputOperator,
		logger:        logger,
		decoder:       helper.NewDecoder(enc),
		splitFunc:     splitFunc,
		scriptContent: scriptContent,
	}, nil
}
