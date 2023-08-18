//go:build !windows
// +build !windows

package host

import (
	"github.com/shirou/gopsutil/v3/process"
)

func (p *processName) getName(proc *process.Process) (string, error) {
	return proc.Name()
}

func (p *processName) setPidNameMap() error {
	return nil
}
