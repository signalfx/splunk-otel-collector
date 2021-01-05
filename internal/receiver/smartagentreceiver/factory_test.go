// Copyright 2021, OpenTelemetry Authors
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

package smartagentreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestCreateReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	receiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, consumertest.NewMetricsNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)
}

func TestCreateReceiverWithInvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{}
	require.Error(t, cfg.validate())

	params := component.ReceiverCreateParams{Logger: zap.NewNop()}
	receiver, err := factory.CreateMetricsReceiver(context.Background(), params, cfg, consumertest.NewMetricsNop())
	assert.Error(t, err)
	assert.Nil(t, receiver)
}
