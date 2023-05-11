package utils

import (
	"context"
	"sync"
	"time"
)

// Debounce0 calls a zero arg function on the trailing edge of every `duration`.
func Debounce0(fn func(), duration time.Duration) (func(), chan<- struct{}) {
	lock := &sync.Mutex{}

	stop := make(chan struct{})
	timer := time.NewTicker(duration)
	callRequested := false

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
				lock.Lock()
				if callRequested {
					fn()
					callRequested = false
				}
				lock.Unlock()
			}
		}
	}()

	return func() {
		lock.Lock()
		callRequested = true
		lock.Unlock()
	}, stop
}

// RunOnInterval the given fn once every interval, starting at the moment the
// function is called.  Returns a function that can be called to stop running
// the function.
func RunOnInterval(ctx context.Context, fn func(), interval time.Duration) {
	timer := time.NewTicker(interval)

	go func() {
		fn()

		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				fn()
			}
		}
	}()
}

// RepeatPolicy repeat behavior for RunOnIntervals Function
type RepeatPolicy int

const (
	// RepeatAll repeats all intervals
	RepeatAll RepeatPolicy = iota
	// RepeatLast repeats only the last interval
	RepeatLast
	// RepeatNone does not repeat
	RepeatNone
)

// RunOnArrayOfIntervals the given function once on the specified intervals, and
// repeat according to the supplied RepeatPolicy.  Please note the function is
// executed after the first interval.  If you want the function executed
// immediately, you should specify a duration of 0 as the first element in the
// intervals array.
func RunOnArrayOfIntervals(ctx context.Context, fn func(), intervals []time.Duration, repeatPolicy RepeatPolicy) {
	// copy intervals
	intvs := append(intervals[:0:0], intervals...)

	// return if the interval list is empty
	if len(intvs) < 1 {
		return
	}

	// set up index and last indice
	index := 0
	lastIndex := len(intvs) - 1

	// initialize timer
	timer := time.NewTimer(intvs[index])
	go func() {
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if index < lastIndex {
					// advance interval
					index++
				} else {
					// evaluate repeat policies
					if repeatPolicy == RepeatNone {
						// execute and return
						fn()
						return
					} else if repeatPolicy == RepeatAll {
						// reset interval
						index = 0
					} // leave index == lastIndex if RepeatLast
				}
				timer.Reset(intvs[index])
				fn()
			}
		}
	}()
}
