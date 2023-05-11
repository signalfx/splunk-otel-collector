//go:build windows
// +build windows

package processlist

import (
	"reflect"
	"testing"
	"time"

	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/neotest"
)

func TestMonitor_Configure(t *testing.T) {
	tests := []struct {
		name       string
		m          *Monitor
		processes  []Win32Process
		threads    []Win32Thread
		cpuPercent map[uint32]uint64
		usernames  map[uint32]string
		want       *event.Event
		wantErr    bool
	}{
		{
			name: "test1",
			m:    &Monitor{Output: neotest.NewTestOutput()},
			processes: []Win32Process{
				{
					Name:           "testProcess1",
					ExecutablePath: pointer.String("C:\\HelloWorld.exe"),
					CommandLine:    pointer.String("HelloWorld.exe"),
					Priority:       8,
					ProcessID:      0,
					Status:         pointer.String(""),
					ExecutionState: pointer.Uint16(0),
					KernelModeTime: 1500,
					PageFileUsage:  1600,
					UserModeTime:   1700,
					WorkingSetSize: 1800,
					VirtualSize:    1900,
				},
				{
					Name:           "testProcess2",
					ExecutablePath: pointer.String("C:\\HelloWorld2.exe"),
					CommandLine:    pointer.String("HelloWorld2.exe"),
					Priority:       8,
					ProcessID:      1,
					Status:         pointer.String(""),
					ExecutionState: pointer.Uint16(0),
					KernelModeTime: 1500,
					PageFileUsage:  1600,
					UserModeTime:   1700,
					WorkingSetSize: 1800,
					VirtualSize:    1900,
				},
			},
			threads: []Win32Thread{
				{
					ThreadState:   3,
					ProcessHandle: "1",
				},
				{
					ThreadState:   5,
					ProcessHandle: "0",
				},
			},
			usernames: map[uint32]string{
				0: "tedMosby",
				1: "barneyStinson",
			},
			want: &event.Event{
				EventType:  "objects.top-info",
				Category:   event.AGENT,
				Dimensions: map[string]string{},
				Properties: map[string]interface{}{
					"message": "{\"t\":\"eJyqVjJQsopWKklN8c0vTqpU0rHQUSrNy87LL89T0jHUMdQx0FEKVtIx0DMwgBBKBgZWBiCWko6Ss1VMjEdqTk5+eH5RTopeakWqUqyOkiHIwKTEorzUyuCSzLzi/DyspgYRZ6oRxNhaQAAAAP//UTMulQ==\",\"v\":\"0.0.30\"}",
				},
			},
		},
		{
			name: "handles nested quotes",
			m:    &Monitor{Output: neotest.NewTestOutput()},
			processes: []Win32Process{
				{
					Name:           "test-proc",
					ExecutablePath: pointer.String("C:\\HelloWorld2\"quoted\".exe"),
					CommandLine:    pointer.String("HelloWorld2.exe"),
					Priority:       8,
					ProcessID:      0,
					Status:         pointer.String(""),
					ExecutionState: pointer.Uint16(0),
					KernelModeTime: 1500,
					PageFileUsage:  1600,
					UserModeTime:   1700,
					WorkingSetSize: 1800,
					VirtualSize:    1900,
				},
			},
			threads: []Win32Thread{
				{
					ThreadState:   5,
					ProcessHandle: "0",
				},
			},
			usernames: map[uint32]string{
				0: "ted\"bud\"Mosby",
			},
			want: &event.Event{
				EventType:  "objects.top-info",
				Category:   event.AGENT,
				Dimensions: map[string]string{},
				Properties: map[string]interface{}{
					"message": "{\"t\":\"eJyqVjJQsopWKklNUU8qTVH3zS9OqlTSsdBRKs3Lzssvz1PSMdQx1DHQUQpW0jHQMzCAEEoGBlYGIJaSjpKzVUyMR2pOTn54flFOipF6YWk+yDS91IpUpdhaQAAAAP///QsbHw==\",\"v\":\"0.0.30\"}",
				},
			},
		},
	}
	for i := range tests {
		origGetAllProcesses := getAllProcesses
		origGetUsername := getUsername
		origGetAllThreads := getAllThreads

		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			getAllProcesses = func() ([]Win32Process, error) {
				return tt.processes, nil
			}
			getUsername = func(id uint32) (string, error) {
				username, ok := tt.usernames[id]
				if !ok {
					t.Error("unable to find username")
				}
				return username, nil
			}
			getAllThreads = func() ([]Win32Thread, error) {
				return tt.threads, nil
			}
			if err := tt.m.Configure(&Config{config.MonitorConfig{IntervalSeconds: 10}}); (err != nil) != tt.wantErr {
				t.Errorf("Monitor.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(3 * time.Second)
			events := tt.m.Output.(*neotest.TestOutput).FlushEvents()
			if len(events) == 0 {
				t.Errorf("events %v != %v", events, tt.want)
				return
			}

			lastEvent := events[len(events)-1]

			w := tt.want
			if lastEvent.EventType != w.EventType ||
				lastEvent.Category != w.Category ||
				!reflect.DeepEqual(lastEvent.Dimensions, w.Dimensions) ||
				!reflect.DeepEqual(lastEvent.Properties, w.Properties) {
				t.Errorf("events %v != %v", lastEvent, tt.want)
				return
			}
		})
		getAllProcesses = origGetAllProcesses
		getUsername = origGetUsername
		getAllThreads = origGetAllThreads
	}
}

func TestMapThreadsToProcess(t *testing.T) {
	type args struct {
		threadList []Win32Thread
	}
	tests := []struct {
		name string
		args args
		want map[string][]uint32
	}{
		{
			name: "check correct mapping",
			args: args{
				threadList: []Win32Thread{
					{ProcessHandle: "2", ThreadState: 3},
				},
			},
			want: map[string][]uint32{
				"2": {3},
			},
		},
		{
			name: "check correct mapping 2",
			args: args{
				threadList: []Win32Thread{
					{ProcessHandle: "1", ThreadState: 3},
					{ProcessHandle: "2", ThreadState: 3},
					{ProcessHandle: "1", ThreadState: 5},
					{ProcessHandle: "1", ThreadState: 5},
				},
			},
			want: map[string][]uint32{
				"1": {3, 5, 5},
				"2": {3},
			},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := mapThreadsToProcess(tt.args.threadList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapThreadsToProcess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusMapping(t *testing.T) {
	type args struct {
		threadStates []uint32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "running process",
			args: args{
				threadStates: []uint32{5, 5, 5, 5, 2, 5},
			},
			want: "R",
		},
		{
			name: "waiting process",
			args: args{
				threadStates: []uint32{5, 5, 5, 5, 5, 5},
			},
			want: "S",
		},
		{
			name: "empty list",
			args: args{
				threadStates: []uint32{},
			},
			want: "",
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := statusMapping(tt.args.threadStates); got != tt.want {
				t.Errorf("statusMapping() = %v, want %v", got, tt.want)
			}
		})
	}
}
