package main

import "testing"

func Test_formatVariable(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"max.cpu", "maxCPU"},
		{"max.foo", "maxFoo"},
		{"cpu.utilization", "cpuUtilization"},
		{"cpu", "cpu"},
		{"max.ip.addr", "maxIPAddr"},
		{"some_metric", "someMetric"},
		{"some-metric", "someMetric"},
		{"Upper.Case", "upperCase"},
		{"max.ip6", "maxIP6"},
		{"max.ip6.idle", "maxIP6Idle"},
		{"node_netstat_IpExt_OutOctets", "nodeNetstatIPExtOutOctets"},
	}
	for _, tt := range tests {
		args := tt.args
		want := tt.want
		t.Run(tt.args, func(t *testing.T) {
			got, err := formatVariable(args)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("formatVariable() = %v, want %v", got, want)
			}
		})
	}
}
