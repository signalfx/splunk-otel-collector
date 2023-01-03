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

package timestampprocessor

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestConfig(t *testing.T) {
	now := time.Now().UTC()
	ts := pcommon.NewTimestampFromTime(now)

	configs, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))

	require.NoError(t, err)
	require.NotNil(t, configs)

	assert.Equal(t, 3, len(configs.ToStringMap()))

	cm, err := configs.Sub(typeStr)
	require.NoError(t, err)
	r0 := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, r0)
	require.NoError(t, err)
	assert.Equal(t, r0, createDefaultConfig())

	cm, err = configs.Sub(fmt.Sprintf("%s/add2h", typeStr))
	require.NoError(t, err)
	r1 := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, r1)
	require.NoError(t, err)
	offset, _ := time.ParseDuration(r1.Offset)
	offset1 := offsetFn(offset)(ts)
	require.Equal(t, now.Add(2*time.Hour), offset1.AsTime())

	cm, err = configs.Sub(fmt.Sprintf("%s/remove3h", typeStr))
	require.NoError(t, err)
	r2 := NewFactory().CreateDefaultConfig().(*Config)
	err = component.UnmarshalConfig(cm, r2)
	require.NoError(t, err)
	offset, _ = time.ParseDuration(r2.Offset)
	offset2 := offsetFn(offset)(ts)
	require.Equal(t, now.Add(-3*time.Hour), offset2.AsTime())
}

func TestOffsetFnZero(t *testing.T) {
	r1 := &Config{
		Offset: "+5h",
	}
	zeroTime := time.Time{}
	require.True(t, zeroTime.IsZero())
	offset, _ := time.ParseDuration(r1.Offset)
	result := offsetFn(offset)(pcommon.Timestamp(uint64(0)))
	require.Equal(t, pcommon.Timestamp(uint64(0)), result)
}
