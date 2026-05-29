//go:build windows

package subproc

import "syscall"

func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
