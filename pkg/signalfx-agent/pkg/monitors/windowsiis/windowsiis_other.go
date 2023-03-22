//go:build !windows
// +build !windows

package windowsiis

import "fmt"

// Configure is the main function for the monitor
func (m *Monitor) Configure(conf *Config) error {
	return fmt.Errorf("this monitor is not implemented on this platform")
}
