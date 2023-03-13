//go:build windows
// +build windows

package host

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/shirou/gopsutil/process"
	"golang.org/x/sys/windows"
)

func (p *processName) getName(proc *process.Process) (string, error) {
	name, ok := p.pidNameMap[proc.Pid]
	if !ok {
		return "", fmt.Errorf("could not find name for PID %v", proc.Pid)
	}
	return name, nil
}

// setPidNameMap fills up the pidNameMap with all processes running on
// the system by iterating through the windows process snapshot
func (p *processName) setPidNameMap() error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer func() {
		_ = windows.CloseHandle(snapshot)
	}()
	var pe32 windows.ProcessEntry32
	pe32.Size = uint32(unsafe.Sizeof(pe32))
	if err = windows.Process32First(snapshot, &pe32); err != nil {
		return err
	}
	for {
		p.pidNameMap[int32(pe32.ProcessID)] = windows.UTF16ToString(pe32.ExeFile[:])
		if err = windows.Process32Next(snapshot, &pe32); err != nil {
			// ERROR_NO_MORE_FILES we reached the end of the snapshot
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				return nil
			}
			break
		}
	}
	return err
}
