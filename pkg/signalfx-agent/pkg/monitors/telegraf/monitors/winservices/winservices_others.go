//go:build !windows
// +build !windows

package winservices

import "fmt"

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	return fmt.Errorf("%s monitor is only supported on Windows", monitorType)
}
