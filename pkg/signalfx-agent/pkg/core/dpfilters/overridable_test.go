package dpfilters

import (
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/assert"
)

func TestOverridableFilters(t *testing.T) {
	t.Run("Exclude based on simple metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization"}, nil)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))
	})

	t.Run("Excludes based on multiple metric names", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization", "memory.utilization"}, nil)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))

		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on regex metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{`/cpu\..*/`}, nil)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on glob metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{`cpu.util*`, "memor*"}, nil)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on dimension name", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"PO"},
		})

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "PO",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Excludes based on dimension name regex", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`/^[A-Z][A-Z]$/`},
		})

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "PO",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Excludes based on dimension presence", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`/.+/`},
		})

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"host": "localhost",
			},
		}))
	})

	t.Run("Excludes based on dimension name glob", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`*O*`},
		})

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "POD",
			},
		}))

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "POD123",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Excludes based on conjunction of both dimensions and metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{"*.utilization"}, map[string][]string{
			"container_name": {"test"},
		})

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "not matching",
			},
		}))

		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))

		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.usage",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Doesn't match if no dimension filter specified", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization"}, nil)
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Doesn't match if no metric name filter specified", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"mycontainer"},
		})
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
	})

	t.Run("Matches against all dimension pairs", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"host":   {"localhost"},
			"system": {"r4"},
		})
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"host": "localhost",
			}}))
		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"host":   "localhost",
				"system": "r4",
			}}))
	})

	t.Run("Negated dim values take precedent", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"*", "!pause", "!/.*idle/"},
		})
		// Shouldn't match when dimension isn't even present
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "pause",
			}}))
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "is_idle",
			}}))
		assert.True(t, f.Matches(&datapoint.Datapoint{
			Metric: "cpu.utilization",
			Dimensions: map[string]string{
				"container_name": "mycontainer",
			}}))
	})
}
