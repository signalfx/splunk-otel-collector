//go:build !windows
// +build !windows

package cpu

import (
	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

func (m *Monitor) times(perCore bool) ([]cpu.TimesStat, error) {
	return cpu.TimesWithContext(hostfs.Context(), perCore)
}
