//go:build linux
// +build linux

package subproc

import (
	"path/filepath"
	"syscall"
)

// The Linux specific process attribute that make the Python runner be in the
// same process group as the agent so they get shutdown together.
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		// This is Linux-specific and will cause collectd to be killed by the OS if
		// the agent dies
		Pdeathsig: syscall.SIGTERM,
	}
}

func defaultPythonBinaryExecutable(bundleDir string) string {
	return filepath.Join(bundleDir, "bin", "python")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
