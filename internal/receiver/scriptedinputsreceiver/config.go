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
	"os"
	"sort"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/decode"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"go.uber.org/zap"
)

const (
	defaultCollectionInterval = "60s"
	// minMaxLogSize is the minimal size which can be used for buffering TCP input
	minMaxLogSize = 64 * 1024
)

var availableScripts = func() []string {
	var s []string
	for sn := range scripts {
		s = append(s, sn)
	}
	sort.Strings(s)
	return s
}()

type Config struct {
	Multiline          split.Config `mapstructure:"multiline,omitempty"`
	ScriptName         string       `mapstructure:"script_name,omitempty"`
	Encoding           string       `mapstructure:"encoding,omitempty"`
	Source             string       `mapstructure:"source"`
	SourceType         string       `mapstructure:"sourcetype"`
	CollectionInterval string       `mapstructure:"collection_interval"`
	helper.InputConfig `mapstructure:",squash"`
	MaxLogSize         helper.ByteSize `mapstructure:"max_log_size,omitempty"`
	interval           time.Duration
	AddAttributes      bool `mapstructure:"add_attributes,omitempty"`
}

func createDefaultConfig() *Config {
	return &Config{
		Encoding:           "utf-8",
		InputConfig:        helper.NewInputConfig(typeStr, typeStr),
		Multiline:          split.Config{},
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
		return fmt.Errorf("unsupported 'script_name' %q. must be one of %v", c.ScriptName, availableScripts)
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
	if isContainer() {
		return nil, fmt.Errorf("scriped inputs receiver must be run directly on host and is not supported in container")
	}

	inputOperator, err := c.InputConfig.Build(logger)
	if err != nil {
		return nil, err
	}

	enc, err := decode.LookupEncoding(c.Encoding)
	if err != nil {
		return nil, err
	}

	// Build multiline
	var splitFunc bufio.SplitFunc
	if c.Multiline.LineStartPattern == "" && c.Multiline.LineEndPattern == "" {
		splitFunc = split.NoSplitFunc(int(c.MaxLogSize))
	} else {
		splitFunc, err = c.Multiline.Func(enc, true, int(c.MaxLogSize))
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
		decoder:       decode.New(enc),
		splitFunc:     splitFunc,
		scriptContent: scriptContent,
	}, nil
}

func isContainer() bool {
	inContainer := os.Getpid() == 1
	for _, p := range []string{
		"/.dockerenv",        // Mounted by dockerd when starting a container by default
		"/run/.containerenv", // Mounted by podman as described here: https://github.com/containers/podman/blob/ecbb52cb478309cfd59cc061f082702b69f0f4b7/docs/source/markdown/podman-run.1.md.in#L31
	} {
		if _, err := os.Stat(p); err == nil {
			inContainer = true
			break
		}
	}
	return inContainer
}
