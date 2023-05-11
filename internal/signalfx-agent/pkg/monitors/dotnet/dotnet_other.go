//go:build !windows
// +build !windows

package dotnet

import "fmt"

// Configure is the monitor
func (m *Monitor) Configure(conf *Config) error {
	return fmt.Errorf("this monitor is not implemented on this platform")
}
