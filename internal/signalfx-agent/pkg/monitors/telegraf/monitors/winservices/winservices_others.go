//go:build !windows

package winservices

import "fmt"

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(_ *Config) error {
	return fmt.Errorf("%s monitor is only supported on Windows", monitorType)
}
