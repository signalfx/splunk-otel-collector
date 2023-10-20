//go:build windows
// +build windows

package subproc

import (
	"path/filepath"
	"syscall"
)

// The Windows specific process attributes
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		//Pdeathsig: syscall.SIGTERM,
	}
}

func defaultPythonBinaryExecutable(bundleDir string) string {
	return filepath.Join(bundleDir, "python", "python.exe")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
