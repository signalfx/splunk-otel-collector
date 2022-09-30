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

package timestamp

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestConfig(t *testing.T) {
	now := time.Now().UTC()
	ts := pcommon.NewTimestampFromTime(now)

	factories, err := componenttest.NopFactories()
	assert.Nil(t, err)

	factory := NewFactory()
	factories.Processors[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Processors), 3)

	r0 := cfg.Processors[config.NewComponentID(typeStr)].(*Config)
	assert.Equal(t, r0, createDefaultConfig())

	r1 := cfg.Processors[config.NewComponentIDWithName(typeStr, "add2h")].(*Config)
	offset1 := r1.offsetFn()(ts)
	require.Equal(t, now.Add(2*time.Hour), offset1.AsTime())

	r2 := cfg.Processors[config.NewComponentIDWithName(typeStr, "remove3h")].(*Config)
	offset2 := r2.offsetFn()(ts)
	require.Equal(t, now.Add(-3*time.Hour), offset2.AsTime())
}

func TestOffsetFnZero(t *testing.T) {
	r1 := &Config{
		Offset: "+5h",
	}
	zeroTime := time.Time{}
	require.True(t, zeroTime.IsZero())
	result := r1.offsetFn()(pcommon.Timestamp(uint64(0)))
	require.Equal(t, pcommon.Timestamp(uint64(0)), result)
}
