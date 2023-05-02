//go:build windows
// +build windows

package processlist

import (
	"fmt"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/mem"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

const (
	// represents the thread is in waiting state
	// ref: https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-thread
	threadWaitingState             = 5
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
	ThreadState   uint32
	ProcessHandle string
}

// PerfProcProcess is a performance process struct used for wmi calls
// https://msdn.microsoft.com/en-us/library/aa394323(v=vs.85).aspx
type PerfProcProcess struct {
	IDProcess            uint32
	PercentProcessorTime uint64
}

type osCache struct {
}

func initOSCache() *osCache {
	return &osCache{}
}

// getAllProcesses retrieves all processes.  It is set as a package variable so we can mock it during testing
var getAllProcesses = func() (ps []Win32Process, err error) {
	err = wmi.Query("select Name, ExecutablePath, CommandLine, CreationDate, Priority, ProcessID, Status, ExecutionState, KernelModeTime, PageFileUsage, UserModeTime, WorkingSetSize, VirtualSize from Win32_Process", &ps)
	return ps, err
}

// getAllThreads retrieves all the threads.  It is set as a package variable so we can mock it during testing
var getAllThreads = func() (threads []Win32Thread, err error) {
	err = wmi.Query("select ThreadState, ProcessHandle from Win32_Thread", &threads)
	return threads, err
}

// getUsername - retrieves a username from an open process handle it is set as a package variable so we can mock it during testing
var getUsername = func(id uint32) (username string, err error) {
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
func ProcessList(conf *Config, cache *osCache, logger logrus.FieldLogger) ([]*TopProcess, error) {
	var procs []*TopProcess

	// Get all processes
	ps, err := getAllProcesses()
	if err != nil {
		return nil, err
	}

	// Get all threads
	threads, err := getAllThreads()
	if err != nil {
		return nil, err
	}

	processMap := mapThreadsToProcess(threads)
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
		status := statusMapping(processMap[fmt.Sprint(p.ProcessID)])
		//example process "3":["root",20,"0",0,0,0,"S",0.0,0.0,"01:28.31","[ksoftirqd/0]"]
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

// Mapping each thread's state to its respective process.
// for example, threadList = []Win32Thread{{ProcessHandle: "1", ThreadState: 3},
// {ProcessHandle: "2", ThreadState: 3},{ProcessHandle: "1", ThreadState: 5},{ProcessHandle: "1", ThreadState: 5},}
// it returns map[string][]uint32{"1": []uint32{3, 5, 5}, "2": []uint32{3},},
func mapThreadsToProcess(threadList []Win32Thread) map[string][]uint32 {
	var processes = make(map[string][]uint32)
	for _, thread := range threadList {
		processes[thread.ProcessHandle] = append(processes[thread.ProcessHandle], thread.ThreadState)
	}
	return processes
}

// Returns the process status depending upon all thread's state.
// if all the threads of a process are in waiting state then it returns "S"(sleeping)
// else it returns "R"(running)
func statusMapping(threadStates []uint32) string {
	if len(threadStates) == 0 {
		return ""
	}

	for _, state := range threadStates {
		if state != threadWaitingState {
			return "R"
		}
	}
	return "S"
}
