package utils

import (
	"strconv"
	"sync"
)

// NewIDGenerator returns a function that will produce, for any given generator
// instance, a unique, non-empty, string value each time it is called.
func NewIDGenerator() func() string {
	lock := sync.Mutex{}
	nextID := 0

	return func() string {
		lock.Lock()
		defer lock.Unlock()

		nextID++
		return strconv.Itoa(nextID)
	}
}
