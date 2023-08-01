// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package procpipe

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateDefaultConfig(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)

	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, builded, "failed to create default config")
}

func TestCreateDefaultWithSmallLogSize(t *testing.T) {
	config := Config{}
	config.ScriptName = "df"
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2
	err := config.Validate()

	assert.Equal(t, err.Error(), "invalid value for parameter 'max_log_size', must be equal to or greater than 65536 bytes")
}

func TestCreateDefaultWithMissingExecFile(t *testing.T) {
	config := Config{}
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2

	err := config.Validate()

	assert.Equal(t, err.Error(), "'exec_file' must be specified")
}

func TestCreateDefaultWithNonEmptyMultiline(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	config.Multiline = helper.MultilineConfig{
		LineStartPattern: "a",
		LineEndPattern:   "",
	}

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)
	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, builded, "failed to create default config")
}

func TestCreateTicker(t *testing.T) {
	ticker, err := createTicker("20s")

	assert.NotNil(t, ticker, "failed to create default config")
	assert.Nil(t, err, "failed to create default config")
}

func TestReadOutput(t *testing.T) {
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"
	config.CollectionInterval = "60s"

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	builded, _ := config.Build(sugar)
	builded.Start(nil)
	builded.Stop()

	assert.NotNil(t, builded, "failed to create default config")
}
