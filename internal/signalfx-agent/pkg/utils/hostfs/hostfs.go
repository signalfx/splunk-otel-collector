package hostfs

import (
	"context"
	"sync"

	"github.com/shirou/gopsutil/v3/common"
)

var (
	lock   sync.RWMutex
	envMap common.EnvMap
)

func SetEnvMap(m common.EnvMap) {
	lock.Lock()
	defer lock.Unlock()
	envMap = m
}

// Context returns a context for gopsutil function calls.
func Context() context.Context {
	lock.RLock()
	defer lock.RUnlock()
	return context.WithValue(context.Background(), common.EnvKey, envMap)
}

// HostProc returns the configured /proc path
func HostProc() string {
	lock.RLock()
	defer lock.RUnlock()
	return envMap[common.HostProcEnvKey]
}

// HostEtc returns the configured /etc path
func HostEtc() string {
	lock.RLock()
	defer lock.RUnlock()
	return envMap[common.HostEtcEnvKey]
}

// HostRun returns the configured /run path
func HostRun() string {
	lock.RLock()
	defer lock.RUnlock()
	return envMap[common.HostRunEnvKey]
}

// HostVar returns the configured /var path
func HostVar() string {
	lock.RLock()
	defer lock.RUnlock()
	return envMap[common.HostVarEnvKey]
}

// HostSys returns the configured host /sys path
func HostSys() string {
	lock.RLock()
	defer lock.RUnlock()
	return envMap[common.HostSysEnvKey]
}
