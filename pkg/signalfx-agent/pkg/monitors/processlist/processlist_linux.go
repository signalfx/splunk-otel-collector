//go:build linux
// +build linux

package processlist

import (
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/prometheus/procfs"
	"github.com/sirupsen/logrus"
)

// A place to hold system info that is assumed not to change (or rarely change)
type osCache struct {
	pageSize int
	uidCache map[string]*user.User
}

func initOSCache() *osCache {
	return &osCache{
		pageSize: os.Getpagesize(),
		uidCache: map[string]*user.User{},
	}
}

// ProcessList takes a snapshot of running processes
func ProcessList(conf *Config, cache *osCache, logger logrus.FieldLogger) ([]*TopProcess, error) {
	var fs procfs.FS
	var err error
	if conf.ProcPath == "" {
		fs, err = procfs.NewDefaultFS()
	} else {
		fs, err = procfs.NewFS(conf.ProcPath)
	}
	if err != nil {
		return nil, err
	}

	procs, err := fs.AllProcs()
	if err != nil {
		return nil, err
	}

	hostMem, _ := fs.Meminfo()

	var out []*TopProcess
	for _, p := range procs {
		stat, err := p.Stat()
		if err != nil {
			continue
		}

		status, err := p.NewStatus()
		if err != nil {
			continue
		}

		cmdLine, _ := p.CmdLine()
		if len(cmdLine) == 0 {
			comm, _ := p.Comm()
			cmdLine = []string{comm}
		}

		st, _ := stat.StartTime()

		username := ""
		uid := status.UIDs[0]
		if uid != "" {
			cachedUser := cache.uidCache[uid]
			if cachedUser != nil {
				username = cachedUser.Username
			} else {
				user, err := user.LookupId(uid)
				if err == nil {
					cache.uidCache[uid] = user
					username = user.Username
				} else if logger != nil {
					logger.WithError(err).Debugf("Could not lookup user id %s for process id %d", uid, p.PID)
				}
			}
		}

		var memPercent float64
		if hostMem.MemTotal != nil {
			memPercent = 100.0 * float64(stat.RSS*cache.pageSize) / float64(*hostMem.MemTotal*1024)
		}

		out = append(out, &TopProcess{
			ProcessID:           p.PID,
			CreatedTime:         time.Unix(int64(st), 0),
			Username:            username,
			Priority:            stat.Priority,
			Nice:                &stat.Nice,
			VirtualMemoryBytes:  uint64(stat.VirtualMemory()),
			WorkingSetSizeBytes: uint64(stat.RSS * cache.pageSize),
			SharedMemBytes:      status.RssShmem + status.RssFile,
			Status:              stat.State,
			MemPercent:          memPercent,
			// gopsutil scales the times to seconds already
			TotalCPUTime: time.Duration(stat.CPUTime() * float64(time.Second)),
			Command:      strings.Join(cmdLine, " "),
		})
	}
	return out, nil
}
