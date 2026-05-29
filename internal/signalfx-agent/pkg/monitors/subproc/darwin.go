//go:build darwin

package subproc

import "syscall"

func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
