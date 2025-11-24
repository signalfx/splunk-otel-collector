package config

import (
	"testing"

	"github.com/signalfx/golib/v3/datapoint" //nolint:staticcheck // SA1019: deprecated package still in use
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFilters(t *testing.T) {
	t.Run("Make single filter properly", func(t *testing.T) {
		f, _ := makeNewFilterSet([]MetricFilter{
			{
				MetricNames: []string{
					"cpu.utilization",
					"memory.utilization",
				},
			},
		})
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Merges two filters properly (ORed together)", func(t *testing.T) {
		f, _ := makeNewFilterSet([]MetricFilter{
			{
				MetricNames: []string{
					"cpu.utilization",
					"memory.utilization",
				},
			},
			{
				MetricNames: []string{
					"disk.utilization",
				},
			},
		})
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "other.utilization"}))
	})

	t.Run("Filters can be overridden within a single filter", func(t *testing.T) {
		f, _ := makeNewFilterSet([]MetricFilter{
			{
				MetricNames: []string{
					"*.utilization",
					"!memory.utilization",
					"!/[a-c].*.utilization/",
				},
			},
			{
				MetricNames: []string{
					"disk.utilization",
				},
			},
		})
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "network.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "other.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))
	})

	t.Run("Filters respect both metric names and dimensions", func(t *testing.T) {
		f, err := makeNewFilterSet([]MetricFilter{
			{
				MetricNames: []string{
					"*.utilization",
					"!memory.utilization",
				},
				Dimensions: map[string]interface{}{
					"env":     []interface{}{"prod", "dev"},
					"service": []interface{}{"db"},
				},
			},
			{
				MetricNames: []string{
					"disk.utilization",
				},
			},
			{
				Dimensions: map[string]interface{}{
					"service": "es",
				},
			},
		})
		require.NoError(t, err)

		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"env": "prod",
			},
		}))

		// No env dimension and metric name negated so not filtered
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "memory.utilization",
			Dimensions: map[string]string{
				"env": "prod",
			},
		}))

		// Metric name is negated
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "memory.utilization",
			Dimensions: map[string]string{
				"env":     "prod",
				"service": "db",
			},
		}))

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"env":     "prod",
				"service": "db",
			},
		}))

		// One dimension missing
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"env": "prod",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))

		// Matches by dimension only
		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "random.metric",
			Dimensions: map[string]string{
				"service": "es",
			},
		}))
	})
}
