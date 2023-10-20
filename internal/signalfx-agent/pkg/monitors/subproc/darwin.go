//go:build darwin
// +build darwin

package subproc

import (
	"path/filepath"
	"syscall"
)

// The Darwin specific process attribute that make the Python runner be in the
// same process group as the agent so they get shutdown together.
func procAttrs() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func defaultPythonBinaryExecutable(bundleDir string) string {
	return filepath.Join(bundleDir, "bin/python")
}

func defaultPythonBinaryArgs(pkgName string) []string {
	return []string{
		"-u",
		"-m",
		pkgName,
	}
}
