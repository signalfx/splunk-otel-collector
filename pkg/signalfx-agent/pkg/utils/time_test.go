package utils

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

type testMonitor struct {
	executions int64
}

func (t *testMonitor) Execute() {
	atomic.AddInt64(&t.executions, 1)
}

func (t *testMonitor) Count() int64 {
	return atomic.LoadInt64(&t.executions)
}

func TestRunOnArrayOfIntervals(t *testing.T) {
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()
	type args struct {
		ctx          context.Context
		monitor      *testMonitor
		intervals    []time.Duration
		repeatPolicy RepeatPolicy
	}
	tests := []struct {
		name       string
		args       args
		comparison func(got int64) bool
		want       string
	}{
		{
			name: "test repeat none",
			args: args{
				ctx:          context.Background(),
				monitor:      &testMonitor{},
				intervals:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
				repeatPolicy: RepeatNone,
			},
			comparison: func(got int64) bool { return got == 4 },
			want:       "equal to 4",
		},
		{
			name: "test repeat last",
			args: args{
				ctx:          context.Background(),
				monitor:      &testMonitor{},
				intervals:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 30 * time.Millisecond},
				repeatPolicy: RepeatLast,
			},
			comparison: func(got int64) bool { return got > 4 },
			want:       "greater than 4",
		},
		{
			name: "test repeat all",
			args: args{
				ctx:          context.Background(),
				monitor:      &testMonitor{},
				intervals:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
				repeatPolicy: RepeatAll,
			},
			comparison: func(got int64) bool { return got > 8 },
			want:       "greater than 8",
		},
		{
			name: "test no interval",
			args: args{
				ctx:          context.Background(),
				monitor:      &testMonitor{},
				intervals:    []time.Duration{},
				repeatPolicy: RepeatAll,
			},
			comparison: func(got int64) bool { return got == 0 },
			want:       "0",
		},
		{
			name: "test closed context",
			args: args{
				ctx:          cancelledContext,
				monitor:      &testMonitor{},
				intervals:    []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
				repeatPolicy: RepeatAll,
			},
			comparison: func(got int64) bool { return got == 0 },
			want:       "0",
		},
	}
	for _, tt := range tests {
		args := tt.args
		comparison := tt.comparison
		t.Run(tt.name, func(t *testing.T) {
			RunOnArrayOfIntervals(args.ctx, args.monitor.Execute, args.intervals, args.repeatPolicy)
			for !comparison(args.monitor.Count()) {
				runtime.Gosched()
			}
			// ensure we don't continue repeating when repeat policy is set to none
			if args.repeatPolicy == RepeatNone {
				time.Sleep(1 * time.Second)
				runtime.Gosched()
				if !comparison(args.monitor.Count()) {
					t.Errorf("repeat none policy violated")
				}
			}
			// ensure that when the # of intervals is 0 nothing is executed
			if len(args.intervals) == 0 {
				time.Sleep(1 * time.Second)
				runtime.Gosched()
				if !comparison(args.monitor.Count()) {
					t.Errorf("empty interval array violated")
				}
			}
		})
	}
}
