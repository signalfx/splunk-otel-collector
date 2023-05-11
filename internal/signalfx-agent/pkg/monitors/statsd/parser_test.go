package statsd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMetrics(t *testing.T) {
	cases := []struct {
		name       string
		raw        string
		converters []*converter
		parsed     statsDMetric
	}{
		{
			"tags",
			"runtime.node.heap.size.by.space:430080|g|#service:svc1,runtime-id:dcd3,space:code_space",
			nil,
			statsDMetric{
				rawMetricName: "runtime.node.heap.size.by.space",
				metricName:    "runtime.node.heap.size.by.space",
				metricType:    "g",
				value:         430080,
				dimensions: map[string]string{
					"service":    "svc1",
					"runtime-id": "dcd3",
					"space":      "code_space",
				},
			},
		},
		{
			"no-tags",
			"runtime.node.heap.size.by.space:43|g",
			nil,
			statsDMetric{
				rawMetricName: "runtime.node.heap.size.by.space",
				metricName:    "runtime.node.heap.size.by.space",
				metricType:    "g",
				value:         43,
				dimensions:    nil,
			},
		},
		{
			"tags-with-converters",
			"cluster.cds_egress_ecommerce-demo-mesh_gateway-vn_tcp_8080.update_success:100|g|#svc:svc2",
			[]*converter{
				{
					pattern: parseFields("cluster.cds_{traffic}_{mesh}_{service}-vn_{}.{action}", nil),
					metric:  parseFields("{traffic}.{action}", nil),
				},
			},
			statsDMetric{
				rawMetricName: "cluster.cds_egress_ecommerce-demo-mesh_gateway-vn_tcp_8080.update_success",
				metricName:    "egress.update_success",
				metricType:    "g",
				value:         100,
				dimensions: map[string]string{
					"svc":     "svc2",
					"traffic": "egress",
					"mesh":    "ecommerce-demo-mesh",
					"service": "gateway",
					"action":  "update_success",
				},
			},
		},
		{
			"converters-without-tags",
			"cluster.cds_egress_ecommerce-demo-mesh_gateway-vn_tcp_8080.update_success:100|g",
			[]*converter{
				{
					pattern: parseFields("cluster.cds_{traffic}_{mesh}_{service}-vn_{}.{action}", nil),
					metric:  parseFields("{traffic}.{action}", nil),
				},
			},
			statsDMetric{
				rawMetricName: "cluster.cds_egress_ecommerce-demo-mesh_gateway-vn_tcp_8080.update_success",
				metricName:    "egress.update_success",
				metricType:    "g",
				value:         100,
				dimensions: map[string]string{
					"traffic": "egress",
					"mesh":    "ecommerce-demo-mesh",
					"service": "gateway",
					"action":  "update_success",
				},
			},
		},
	}

	for i := range cases {
		tt := cases[i]
		t.Run(tt.name, func(t *testing.T) {
			sl := &statsDListener{
				converters: tt.converters,
				prefix:     "",
			}
			sm := sl.parseMetrics([]string{tt.raw})
			require.Equal(t, 1, len(sm))
			require.Equal(t, tt.parsed, *sm[0])
		})
	}
}

func TestParseFields(t *testing.T) {
	cases := []struct {
		pattern        string
		substrs        []string
		startWithField bool
		expectNil      bool
	}{
		{
			"metric.count",
			[]string{"metric.count"},
			false,
			false,
		},
		{
			"metric.count{",
			nil,
			false,
			true,
		},
		{
			"{metric.count",
			nil,
			false,
			true,
		},
		{
			"metric.count}",
			nil,
			false,
			true,
		},
		{
			"{{metric.count}}",
			nil,
			false,
			true,
		},
		{
			"{metric.count}",
			[]string{"metric.count"},
			true,
			false,
		},
		{
			"cluster.cds_{traffic}_{mesh}_{service}-vn_{}.{action}",
			[]string{"cluster.cds_", "traffic", "_", "mesh", "_", "service", "-vn_", "", ".", "action"},
			false,
			false,
		},
		{
			"cluster.cds_{traffic}_{mesh}_{service}-vn_{}.{action}-prod",
			[]string{"cluster.cds_", "traffic", "_", "mesh", "_", "service", "-vn_", "", ".", "action", "-prod"},
			false,
			false,
		},
		{
			"{cluster}.cds_{traffic}_{mesh}_{service}-vn_{}.{action}",
			[]string{"cluster", ".cds_", "traffic", "_", "mesh", "_", "service", "-vn_", "", ".", "action"},
			true,
			false,
		},
		{
			"{cluster}.cds_{traffic}_{mesh}_{service}-vn_{}.{action}-dev",
			[]string{"cluster", ".cds_", "traffic", "_", "mesh", "_", "service", "-vn_", "", ".", "action", "-dev"},
			true,
			false,
		},
		{
			// Cannot have back-to-back patterns
			"{cluster}.cds_{traffic}{mesh}_{service}-vn_{}.{action}",
			nil,
			false,
			true,
		},
	}
	for i := range cases {
		tt := cases[i]
		t.Run(tt.pattern, func(t *testing.T) {
			fp := parseFields(tt.pattern, nil)
			if tt.expectNil {
				require.Nil(t, fp)
				return
			}

			require.Equal(t, fieldPattern{substrs: tt.substrs, startWithField: tt.startWithField}, *fp)
		})
	}
}
