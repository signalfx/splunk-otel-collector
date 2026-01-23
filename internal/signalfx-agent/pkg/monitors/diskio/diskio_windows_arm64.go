//go:build windows && arm64

package diskio

import "fmt"

type Monitor struct {
	cancel func()
}

// Configure is the main function for the monitor
func (m *Monitor) Configure(_ *Config) error {
	m.cancel = func() {}

	return fmt.Errorf("this monitor is not implemented on this platform")
}
