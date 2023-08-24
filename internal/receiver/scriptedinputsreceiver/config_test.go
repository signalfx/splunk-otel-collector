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
	"path"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.uber.org/zap"
)

func tmpScript(t *testing.T) {
	scripts["aasd"] = "tmp"
	t.Cleanup(func() {
		delete(scripts, "aasd")
	})
}

func TestValidConfig(t *testing.T) {
	configs, err := confmaptest.LoadConf(path.Join(".", "testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, configs)

	assert.Equal(t, 1, len(configs.ToStringMap()))

	cm, err := configs.Sub("scripted_inputs/cpu")
	require.NoError(t, err)

	cfg := createDefaultConfig()
	err = component.UnmarshalConfig(cm, cfg)
	require.NoError(t, err)

	require.Equal(t, &Config{
		InputConfig: helper.InputConfig{
			AttributerConfig: helper.AttributerConfig{
				Attributes: map[string]helper.ExprStringConfig{},
			},
			IdentifierConfig: helper.IdentifierConfig{
				Resource: map[string]helper.ExprStringConfig{},
			},
			WriterConfig: helper.WriterConfig{
				BasicConfig: helper.BasicConfig{
					OperatorID:   "scripted_inputs",
					OperatorType: "scripted_inputs",
				}, OutputIDs: []string(nil)},
		},
		Multiline: helper.MultilineConfig{
			LineStartPattern: "",
			LineEndPattern:   "",
		},
		ScriptName:         "cpu",
		Encoding:           helper.EncodingConfig{Encoding: "utf-8"},
		Source:             "",
		SourceType:         "",
		CollectionInterval: "60s",
		MaxLogSize:         1048576,
		AddAttributes:      false,
		interval:           0,
	}, cfg)

	require.NoError(t, cfg.Validate())
}

func TestCreateConfig(t *testing.T) {
	tmpScript(t)
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	built, err := config.Build(sugar)
	require.NoError(t, err)

	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, built, "failed to create default config")
}

func TestCreateWithSmallLogSize(t *testing.T) {
	config := Config{}
	config.ScriptName = "df"
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2
	err := config.Validate()

	assert.Equal(t, err.Error(), "invalid value for parameter 'max_log_size', must be equal to or greater than 65536 bytes")
}

func TestCreateWithMissingExecFile(t *testing.T) {
	config := Config{}
	config.OperatorType = "test-operator"
	config.MaxLogSize = 2

	err := config.Validate()

	assert.Equal(t, err.Error(), "'script_name' must be specified")
}

func TestCreateWithNonEmptyMultiline(t *testing.T) {
	tmpScript(t)
	config := Config{}
	config.ScriptName = "aasd"
	config.OperatorType = "test-operator"

	config.Multiline = helper.MultilineConfig{
		LineStartPattern: "a",
		LineEndPattern:   "",
	}

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	built, err := config.Build(sugar)
	require.NoError(t, err)
	assert.NotNil(t, config, "failed to create default config")
	assert.NotNil(t, built, "failed to create default config")
}
