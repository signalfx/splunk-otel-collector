// Copyright  Splunk, Inc.
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

package rollingspanlatencyprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor/internal/metadata"
)

func TestFactory_Type(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, metadata.Type, factory.Type())
}

func TestFactory_CreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	assert.NotNil(t, cfg)

	c := cfg.(*Config)
	assert.Equal(t, defaultConfig(), *c)
}

func TestFactory_CreateTracesProcessor(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	tp, err := factory.CreateTraces(
		context.Background(),
		processortest.NewNopSettings(metadata.Type),
		cfg,
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	assert.NotNil(t, tp)
}

func TestFactory_CreateTracesProcessor_InvalidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	c := cfg.(*Config)
	c.HalfLife = 0 // invalid

	_, err := factory.CreateTraces(
		context.Background(),
		processortest.NewNopSettings(metadata.Type),
		cfg,
		consumertest.NewNop(),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidHalfLife)
}
