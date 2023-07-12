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
	"io"
	"os/exec"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

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

// Commander can start/stop/restart the shell executable and also watch for a signal
// for the shell process to finish.
type Commander struct {
	stdout       io.Writer
	logger       *zap.Logger
	cmd          *exec.Cmd
	doneCh       chan struct{}
	waitCh       chan struct{}
	execFilePath string
	args         []string
	running      int64
}

func NewCommander(logger *zap.Logger, execFilePath string, stdout io.Writer, args ...string) (*Commander, error) {
	return &Commander{
		execFilePath: execFilePath,
		logger:       logger,
		args:         args,
		stdout:       stdout,
	}, nil
}
