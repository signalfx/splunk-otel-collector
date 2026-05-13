package dpfilters

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/stretchr/testify/assert"
)

var (
	cpu    pmetric.Metric
	memory pmetric.Metric
	disk   pmetric.Metric
)

func init() {
	cpu = pmetric.NewMetric()
	cpu.SetName("cpu.utilization")
	cpu.SetEmptyGauge()
	memory = pmetric.NewMetric()
	memory.SetName("memory.utilization")
	memory.SetEmptyGauge()
	disk = pmetric.NewMetric()
	disk.SetName("disk.utilization")
	disk.SetEmptyGauge()
}

func TestOverridableFilters(t *testing.T) {
	t.Run("Exclude based on simple metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization"}, nil)
		assert.True(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))
		assert.False(t, f.MatchesMetricDataPoint(memory.Name(), pcommon.NewMap()))
	})

	t.Run("Excludes based on multiple metric names", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization", "memory.utilization"}, nil)

		assert.True(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))
		assert.True(t, f.MatchesMetricDataPoint(memory.Name(), pcommon.NewMap()))
		assert.False(t, f.MatchesMetricDataPoint(disk.Name(), pcommon.NewMap()))
	})

	t.Run("Excludes based on regex metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{`/cpu\..*/`}, nil)
		assert.True(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))

		assert.False(t, f.MatchesMetricDataPoint(disk.Name(), pcommon.NewMap()))
	})

	t.Run("Excludes based on glob metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{`cpu.util*`, "memor*"}, nil)
		assert.True(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))
		assert.True(t, f.MatchesMetricDataPoint(memory.Name(), pcommon.NewMap()))

		assert.False(t, f.MatchesMetricDataPoint(disk.Name(), pcommon.NewMap()))
	})

	t.Run("Excludes based on dimension name", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"PO"},
		})

		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "PO")
		assert.True(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
		m2 := pmetric.NewMetric()
		m2.SetName("disk.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")
		assert.False(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Excludes based on dimension name regex", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`/^[A-Z][A-Z]$/`},
		})

		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "PO")
		assert.True(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
		m2 := pmetric.NewMetric()
		m2.SetName("disk.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")
		assert.False(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Excludes based on dimension presence", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`/.+/`},
		})

		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")
		assert.True(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
		m2 := pmetric.NewMetric()
		m2.SetName("cpu.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("host", "localhost")
		assert.False(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Excludes based on dimension name glob", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {`*O*`},
		})
		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "POD")

		assert.True(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))

		m2 := pmetric.NewMetric()
		m2.SetName("cpu.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "POD123")
		assert.True(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))

		m3 := pmetric.NewMetric()
		m3.SetName("disk.utilization")
		m3.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")

		assert.False(t, f.MatchesMetricDataPoint(m3.Name(), m3.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Excludes based on conjunction of both dimensions and metric name", func(t *testing.T) {
		f, _ := NewOverridable([]string{"*.utilization"}, map[string][]string{
			"container_name": {"test"},
		})

		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "not_matching")

		assert.False(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))

		m2 := pmetric.NewMetric()
		m2.SetName("disk.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")

		assert.True(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))

		m3 := pmetric.NewMetric()
		m3.SetName("disk.usage")
		m3.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")

		assert.False(t, f.MatchesMetricDataPoint(m3.Name(), m3.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Doesn't match if no dimension filter specified", func(t *testing.T) {
		f, _ := NewOverridable([]string{"cpu.utilization"}, nil)
		m := pmetric.NewMetric()
		m.SetName("disk.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "test")
		assert.False(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Doesn't match if no metric name filter specified", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"mycontainer"},
		})
		assert.False(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))
	})

	t.Run("Matches against all dimension pairs", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"host":   {"localhost"},
			"system": {"r4"},
		})
		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("host", "localhost")
		assert.False(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
		m2 := pmetric.NewMetric()
		m2.SetName("cpu.utilization")
		attrs := m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes()
		attrs.PutStr("host", "localhost")
		attrs.PutStr("system", "r4")

		assert.True(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))
	})

	t.Run("Negated dim values take precedent", func(t *testing.T) {
		f, _ := NewOverridable(nil, map[string][]string{
			"container_name": {"*", "!pause", "!/.*idle/"},
		})
		// Shouldn't match when dimension isn't even present
		assert.False(t, f.MatchesMetricDataPoint(cpu.Name(), pcommon.NewMap()))
		m := pmetric.NewMetric()
		m.SetName("cpu.utilization")
		m.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "pause")
		assert.False(t, f.MatchesMetricDataPoint(m.Name(), m.Gauge().DataPoints().At(0).Attributes()))
		m2 := pmetric.NewMetric()
		m2.SetName("cpu.utilization")
		m2.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "is_idle")
		assert.False(t, f.MatchesMetricDataPoint(m2.Name(), m2.Gauge().DataPoints().At(0).Attributes()))
		m3 := pmetric.NewMetric()
		m3.SetName("cpu.utilization")
		m3.SetEmptyGauge().DataPoints().AppendEmpty().Attributes().PutStr("container_name", "mycontainer")
		assert.True(t, f.MatchesMetricDataPoint(m3.Name(), m3.Gauge().DataPoints().At(0).Attributes()))
	})
}
