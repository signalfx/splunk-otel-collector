//go:build linux

package subproc

import "syscall"

func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
}
