//go:build !windows
// +build !windows

package winperfcounters

import "fmt"

// Configure the monitor
func (m *Monitor) Configure(conf *Config) error {
	return fmt.Errorf("%s monitor is only supported on Windows", monitorType)
}
