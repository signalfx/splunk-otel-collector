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
	"testing"

	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestOutput(t *testing.T) {
	output := NewOutput(Config{}, consumertest.NewMetricsNop(), zap.NewNop())
	output.AddDatapointExclusionFilter(dpfilters.DatapointFilter(nil))
	assert.Empty(t, output.EnabledMetrics())
	assert.False(t, output.HasEnabledMetricInGroup(""))
	assert.False(t, output.HasAnyExtraMetrics())
	assert.NotSame(t, &output, output.Copy())
	output.SendDatapoints()
	output.SendEvent(new(event.Event))
	output.SendSpans()
	output.SendDimensionUpdate(new(types.Dimension))
	output.AddExtraDimension("", "")
	output.RemoveExtraDimension("")
	output.AddExtraSpanTag("", "")
	output.RemoveExtraSpanTag("")
	output.AddDefaultSpanTag("", "")
	output.RemoveDefaultSpanTag("")
}

func TestExtraDimensions(t *testing.T) {
	output := NewOutput(Config{}, consumertest.NewMetricsNop(), zap.NewNop())
	assert.Empty(t, output.extraDimensions)

	output.RemoveExtraDimension("not_a_known_dimension_name")

	output.AddExtraDimension("a_dimension_name", "a_value")
	assert.Equal(t, "a_value", output.extraDimensions["a_dimension_name"])

	cp, ok := output.Copy().(*Output)
	require.True(t, ok)
	assert.Equal(t, "a_value", cp.extraDimensions["a_dimension_name"])

	cp.RemoveExtraDimension("a_dimension_name")
	assert.Empty(t, cp.extraDimensions["a_dimension_name"])
	assert.Equal(t, "a_value", output.extraDimensions["a_dimension_name"])

	cp.AddExtraDimension("another_dimension_name", "another_dimension_value")
	assert.Equal(t, "another_dimension_value", cp.extraDimensions["another_dimension_name"])
	assert.Empty(t, output.extraDimensions["another_dimension_name"])
}
