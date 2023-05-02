//go:build !windows
// +build !windows

package cpu

import (
	"github.com/shirou/gopsutil/cpu"
)

func (m *Monitor) times(perCore bool) ([]cpu.TimesStat, error) {
	return cpu.Times(perCore)
}
