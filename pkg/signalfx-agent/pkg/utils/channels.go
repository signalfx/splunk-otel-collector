package utils

// IsSignalChanClosed returns whether a channel is closed that is used only for
// the sake of sending a single singal.  The channel should never be sent any
// actual values, but should only be closed to tell other goroutines to stop.
func IsSignalChanClosed(ch <-chan struct{}) bool {
	if ch == nil {
		return true
	}

	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}
