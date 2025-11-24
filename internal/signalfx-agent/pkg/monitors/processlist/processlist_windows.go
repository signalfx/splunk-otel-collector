//go:build windows

package processlist

import (
	"fmt"
	"strconv"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

const (
	processQueryLimitedInformation = 0x00001000
)

// Win32Process is a WMI struct used for WMI calls
// https://docs.microsoft.com/en-us/windows/desktop/CIMWin32Prov/win32-process
type Win32Process struct {
	Name           string
	ExecutablePath *string
	CommandLine    *string
	CreationDate   time.Time
	Priority       uint32
	ProcessID      uint32
	Status         *string
	ExecutionState *uint16
	KernelModeTime uint64
	PageFileUsage  uint32
	UserModeTime   uint64
	WorkingSetSize uint64
	VirtualSize    uint64
}

// Win32Thread is a WMI struct used for WMI calls
// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-thread
type Win32Thread struct {
	ProcessHandle string
}

// PerfProcProcess is a performance process struct used for wmi calls
// https://msdn.microsoft.com/en-us/library/aa394323(v=vs.85).aspx
type PerfProcProcess struct {
	IDProcess            uint32
	PercentProcessorTime uint64
}

type osCache struct{}

func initOSCache() *osCache {
	return &osCache{}
}

// getAllProcesses retrieves all processes.
func getAllProcesses() (ps []Win32Process, err error) {
	err = wmi.Query("select Name, ExecutablePath, CommandLine, CreationDate, Priority, ProcessID, Status, ExecutionState, KernelModeTime, PageFileUsage, UserModeTime, WorkingSetSize, VirtualSize from Win32_Process", &ps)
	return ps, err
}

// getProcessesWithNonWaitingThreads retrieves all processes with non-waiting threads.
func getProcessesWithNonWaitingThreads() (map[uint32]struct{}, error) {
	var nonWaitingThreads []Win32Thread
	// ThreadState equals 5 means the thread is in a waiting state.
	// ref: https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-thread
	err := wmi.Query("select ProcessHandle from Win32_Thread where ThreadState<>5", &nonWaitingThreads)
	if err != nil {
		return make(map[uint32]struct{}), err
	}

	processWithNonWaitingThreads := make(map[uint32]struct{}, len(nonWaitingThreads))
	for _, nonWaitingThread := range nonWaitingThreads {
		val, err := strconv.ParseUint(nonWaitingThread.ProcessHandle, 10, 32)
		if err == nil {
			pid := uint32(val) //nolint:gosec
			if _, ok := processWithNonWaitingThreads[pid]; !ok {
				processWithNonWaitingThreads[pid] = struct{}{}
			}
		}
	}
	return processWithNonWaitingThreads, nil
}

// getUsername - retrieves a username from an open process handle.
func getUsername(id uint32) (username string, err error) {
	// open the process handle and collect any information that requires it
	var h windows.Handle
	if h, err = windows.OpenProcess(processQueryLimitedInformation, false, id); err != nil {
		err = fmt.Errorf("unable to open process handle. %v", err)
		return username, err
	}
	// Deferring CloseHandle(h) before it is set require a reference on the closure, avoid that
	// by only deferring it only after the handle is successfully opened.
	defer func(h windows.Handle) { _ = windows.CloseHandle(h) }(h)

	// the windows api docs suggest that windows.TOKEN_READ is a super set of windows.TOKEN_QUERY,
	// but in practice windows.TOKEN_READ seems to be less permissive for the admin user
	var token windows.Token
	err = windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token)
	if err != nil {
		err = fmt.Errorf("unable to retrieve process token. %v", err)
		return username, err
	}
	// Do not defer token.Close right after declaration, only after the token is successfully set
	// since token.Close is a value receiver.
	defer token.Close()

	// extract the user from the process token
	user, err := token.GetTokenUser()
	if err != nil {
		err = fmt.Errorf("unable to get token user. %v", err)
		return username, err
	}

	// extract the username and domain from the user
	userid, domain, _, err := user.User.Sid.LookupAccount("")
	if err != nil {
		err = fmt.Errorf("unable to look up user account from Sid %v", err)
	}
	username = fmt.Sprintf("%s\\%s", domain, userid)

	return username, err
}

// ProcessList takes a snapshot of running processes
func ProcessList(_ *Config, _ *osCache, logger logrus.FieldLogger) ([]*TopProcess, error) {
	var procs []*TopProcess

	// Get all processes
	ps, err := getAllProcesses()
	if err != nil {
		return nil, err
	}

	// Get a map of processes with running threads
	processWithNonWaitingThreads, err := getProcessesWithNonWaitingThreads()
	if err != nil && logger != nil {
		logger.Debugf("Unable to collect non waiting threads. %v", err)
	}

	// iterate over each process and build an entry for the process list
	for _, p := range ps {
		username, err := getUsername(p.ProcessID)
		if err != nil && logger != nil {
			logger.Debugf("Unable to collect username for process %v. %v", p, err)
		}

		totalTime := time.Duration(float64(p.UserModeTime+p.KernelModeTime) * 100) // 100 ns units

		// Memory Percent
		var memPercent float64
		if systemMemory, err := mem.VirtualMemory(); err == nil {
			memPercent = 100 * float64(p.WorkingSetSize) / float64(systemMemory.Total)
		} else if logger != nil {
			logger.WithError(err).Error("Unable to collect system memory total")
		}

		// some windows processes do not have an executable path, but they do have a name
		command := *p.ExecutablePath
		if command == "" {
			command = p.Name
		}

		// update process status
		status := "S"
		if _, ok := processWithNonWaitingThreads[p.ProcessID]; ok {
			status = "R"
		}

		// example process "3":["root",20,"0",0,0,0,"S",0.0,0.0,"01:28.31","[ksoftirqd/0]"]
		procs = append(procs, &TopProcess{
			ProcessID:           int(p.ProcessID),
			CreatedTime:         p.CreationDate,
			Username:            username,
			Priority:            int(p.Priority),
			Nice:                nil, // nice value is not available on windows
			VirtualMemoryBytes:  p.VirtualSize,
			WorkingSetSizeBytes: p.WorkingSetSize,
			SharedMemBytes:      0,
			Status:              status,
			MemPercent:          memPercent,
			TotalCPUTime:        totalTime,
			Command:             command,
		})
	}
	return procs, nil
}
