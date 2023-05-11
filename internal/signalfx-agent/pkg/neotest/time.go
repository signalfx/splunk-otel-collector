package neotest

import "time"

// FixedTime is always at 2016-01-05T15:04:05-06:00
var FixedTime func() time.Time

func init() {
	if fixedTime, err := time.Parse(time.RFC3339, "2016-01-05T15:04:05-06:00"); err != nil {
		panic("unable to parse time")
	} else {
		FixedTime = func() time.Time { return fixedTime }
	}
}

// PinnedNow returns a function that is compatible with time.Now that always
// returns the given t Time.
func PinnedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// AdvancedNow returns a now time.Now-compatible function that returns the time
// returned from oldNow advanced by the given Duration d.
func AdvancedNow(oldNow func() time.Time, d time.Duration) func() time.Time {
	return PinnedNow(oldNow().Add(d))
}
