//go:build !windows
// +build !windows

package host

import (
	"github.com/shirou/gopsutil/process"
)

func (p *processName) getName(proc *process.Process) (string, error) {
	return proc.Name()
}

func (p *processName) setPidNameMap() error {
	return nil
}
