//go:build !windows
// +build !windows

package aspdotnet

import "fmt"

// Configure is the monitor
func (m *Monitor) Configure(_ *Config) error {
	return fmt.Errorf("this monitor is not implemented on this platform")
}
