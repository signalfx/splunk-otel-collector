// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rollingspanlatencyprocessor // import "github.com/signalfx/splunk-otel-collector/pkg/processor/rollingspanlatencyprocessor"

import (
	"math"
	"sync"
	"time"
)

// spanStats maintains an exponentially weighted mean and variance for a single
// span name. The decay factor alpha is derived from the configured half-life:
//
//	alpha = 1 - exp(-ln(2) * dt / halfLife)
//
// where dt is the elapsed time since the last observation. This makes the
// effective weight of any sample halve every halfLife regardless of observation
// frequency.
type spanStats struct {
	mu       sync.Mutex
	mean     float64
	variance float64 // Welford's online EWMA variance
	lastSeen time.Time
	count    int64 // total observations; used to gate labeling until warm
}

// update incorporates a new duration sample (nanoseconds) at the given wall
// time, returning the current mean and stddev after the update.
func (s *spanStats) update(durationNs float64, now time.Time, halfLife time.Duration) (mean, stddev float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alpha := s.decayAlpha(now, halfLife)
	s.lastSeen = now
	s.count++

	if s.count == 1 {
		s.mean = durationNs
		s.variance = 0
		return s.mean, 0
	}

	diff := durationNs - s.mean
	s.mean += alpha * diff
	// EWMA variance: V_t = (1-α)*(V_{t-1} + α*diff²)
	s.variance = (1 - alpha) * (s.variance + alpha*diff*diff)

	return s.mean, math.Sqrt(s.variance)
}

// snapshot returns the current mean and stddev without updating.
func (s *spanStats) snapshot() (mean, stddev float64, count int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mean, math.Sqrt(s.variance), s.count
}

// idleSince returns the time of the most recent observation, used by the
// eviction sweep to determine whether the entry has gone stale.
func (s *spanStats) idleSince() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSeen
}

// decayAlpha computes alpha based on elapsed time since lastSeen. For the very
// first call (lastSeen.IsZero) it returns 1.0 so the first sample seeds the
// mean directly.
func (s *spanStats) decayAlpha(now time.Time, halfLife time.Duration) float64 {
	if s.lastSeen.IsZero() {
		return 1.0
	}
	dt := now.Sub(s.lastSeen).Seconds()
	hl := halfLife.Seconds()
	// alpha = 1 - exp(-ln2 * dt/halfLife)
	return 1.0 - math.Exp(-math.Ln2*dt/hl)
}
