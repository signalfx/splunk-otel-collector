package dpfilters

import (
	"testing"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {
	t.Run("Exclude based on simple metric name", func(t *testing.T) {
		f, _ := New("", []string{"cpu.utilization"}, nil, false)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))
	})

	t.Run("Excludes based on multiple metric names", func(t *testing.T) {
		f, _ := New("", []string{"cpu.utilization", "memory.utilization"}, nil, false)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))

		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on regex metric name", func(t *testing.T) {
		f, _ := New("", []string{`/cpu\..*/`}, nil, false)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on glob metric name", func(t *testing.T) {
		f, _ := New("", []string{`cpu.util*`, "memor*"}, nil, false)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "memory.utilization"}))

		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "disk.utilization"}))
	})

	t.Run("Excludes based on dimension name", func(t *testing.T) {
		f, _ := New("", nil, map[string][]string{
			"container_name": {"PO"},
		}, false)

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
		f, _ := New("", nil, map[string][]string{
			"container_name": {`/^[A-Z][A-Z]$/`},
		}, false)

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
		f, _ := New("", nil, map[string][]string{
			"container_name": {`/.+/`},
		}, false)

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
		f, _ := New("", nil, map[string][]string{
			"container_name": {`*O*`},
		}, false)

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
		f, _ := New("", []string{"*.utilization"}, map[string][]string{
			"container_name": {"test"},
		}, false)

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
		f, _ := New("", []string{"cpu.utilization"}, nil, false)
		assert.False(t, f.Matches(&datapoint.Datapoint{
			Metric: "disk.utilization",
			Dimensions: map[string]string{
				"container_name": "test",
			},
		}))
	})

	t.Run("Doesn't match if no metric name filter specified", func(t *testing.T) {
		f, _ := New("", nil, map[string][]string{
			"container_name": {"mycontainer"},
		}, false)
		assert.False(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
	})

	t.Run("Matches against all dimension pairs", func(t *testing.T) {
		f, _ := New("", nil, map[string][]string{
			"host":   {"localhost"},
			"system": {"r4"},
		}, false)
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

	t.Run("Matches negated filters", func(t *testing.T) {
		f, _ := New("", nil, map[string][]string{
			"container_name": {"mycontainer"},
		}, true)
		assert.True(t, f.Matches(&datapoint.Datapoint{Metric: "cpu.utilization"}))
	})
}
