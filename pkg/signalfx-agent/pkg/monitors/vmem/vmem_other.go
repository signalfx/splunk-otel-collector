//go:build !windows && !linux
// +build !windows,!linux

package vmem

import "fmt"

// Configure is the main function of the monitor, it will report host metadata
// on a varied interval
func (m *Monitor) Configure(conf *Config) error {
	return fmt.Errorf("this monitor is not implemented on this platform")
}
