//go:build windows
// +build windows

package cpu

import (
	"fmt"
	"unsafe"

	"github.com/shirou/gopsutil/cpu"
	"golang.org/x/sys/windows"
)

const (
	// converts 100ns to computer jiffies
	hundredNSToJiffy = 0.00001

	// systemProcessorPerformanceInformationClass information class to query with NTQuerySystemInformation
	// https://processhacker.sourceforge.io/doc/ntexapi_8h.html#ad5d815b48e8f4da1ef2eb7a2f18a54e0
	systemProcessorPerformanceInformationClass = 8

	// size of systemProcessorPerformanceInfoSize in memory
	systemProcessorPerformanceInfoSize = uint32(unsafe.Sizeof(systemProcessorPerformanceInformation{}))
)

// set gopsutil function to package variable for easier testing
var (
	gopsutilTimes = cpu.Times

	// Windows API DLL
	ntdll        = windows.NewLazySystemDLL("Ntdll.dll")
	ntdllLoadErr = ntdll.Load() // attempt to load the system dll and store the err for reference

	// Windows API Proc
	// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation
	procNtQuerySystemInformation        = ntdll.NewProc("NtQuerySystemInformation")
	procNtQuerySystemInformationLoadErr = procNtQuerySystemInformation.Find() // attempt to find the proc and store the error if needed
)

// SYSTEM_PROCESSOR_PERFORMANCE_INFORMATION
// defined in windows api doc with the following
// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation#system_processor_performance_information
// additional fields documented here
// https://www.geoffchappell.com/studies/windows/km/ntoskrnl/api/ex/sysinfo/processor_performance.htm
type systemProcessorPerformanceInformation struct {
	IdleTime       int64 // idle time in 100ns (this is not a filetime).
	KernelTime     int64 // kernel time in 100ns.  kernel time includes idle time. (this is not a filetime).
	UserTime       int64 // usertime in 100ns (this is not a filetime).
	DpcTime        int64 // dpc time in 100ns (this is not a filetime).
	InterruptTime  int64 // interrupt time in 100ns
	InterruptCount uint32
}

// converts the SYSTEM_PROCESSOR_PERFORMANCE_INFORMATION struct to a gopsutil cpu.TimesStat
func systemProcessorPerformanceInfoToCPUTimesStat(core int, s *systemProcessorPerformanceInformation) cpu.TimesStat {
	return cpu.TimesStat{
		CPU:    fmt.Sprintf("%d", core),
		Idle:   float64(s.IdleTime) * hundredNSToJiffy,
		System: float64(s.KernelTime-s.IdleTime) * hundredNSToJiffy,
		User:   float64(s.UserTime) * hundredNSToJiffy,
		Irq:    float64(s.InterruptTime) * hundredNSToJiffy,
	}
}

// ntQuerySystemInformation gets percore cpu time information using the ntQuerySystemInformation windows api function.
// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation
// According to the windows documentation this is owned by the kernel and could go away in the future.
// However it has been around a long time and the particular method we're using on the returned
// NtQuerySystemInformation has no recommended alternative yet.  If this ever breaks in future Windows
// versions, look at the help doc on ntquerysysteminformation and see if they've created an alternate
// api function to retrieve per core information.
func ntQuerySystemInformation() ([]cpu.TimesStat, error) {
	var coreInfo []cpu.TimesStat
	// ensure dll loaded
	if ntdllLoadErr != nil {
		return coreInfo, fmt.Errorf("failed to load ntdll dll. %s", ntdllLoadErr.Error())
	}
	// ensure proc found
	if procNtQuerySystemInformationLoadErr != nil {
		return coreInfo, fmt.Errorf("failed to find NtQuerySystemInformation proc. %s", procNtQuerySystemInformationLoadErr.Error())
	}
	// Make maxResults large for safety.
	// We can't invoke the api call with a results array that's too small.
	// If we have more than 2056 cores on a single host, then it's probably the future.
	maxBuffer := 2056
	// buffer for results from the windows proc
	resultBuffer := make([]systemProcessorPerformanceInformation, maxBuffer)
	// size of the buffer in memory
	bufferSize := uintptr(systemProcessorPerformanceInfoSize) * uintptr(maxBuffer)
	// size of the returned response
	var retSize uint32

	// Invoke windows api proc.
	// The returned err from the windows dll proc will always be non-nil even when successful.
	// See https://godoc.org/golang.org/x/sys/windows#LazyProc.Call for more information
	retCode, _, err := procNtQuerySystemInformation.Call(
		systemProcessorPerformanceInformationClass, // System Information Class -> SystemProcessorPerformanceInformation
		uintptr(unsafe.Pointer(&resultBuffer[0])),  // pointer to first element in result buffer
		bufferSize,                        // size of the buffer in memory
		uintptr(unsafe.Pointer(&retSize)), // pointer to the size of the returned results the windows proc will set this
	)

	// check return code for errors
	if retCode != 0 {
		return coreInfo, err
	}

	// calculate the number of returned elements based on the returned size
	numReturnedElements := retSize / systemProcessorPerformanceInfoSize

	// trim results to the number of returned elements
	resultBuffer = resultBuffer[:numReturnedElements]

	// convert all of the system processor information to gopsutil timesstats
	coreInfo = make([]cpu.TimesStat, 0, numReturnedElements)
	for core := range resultBuffer {
		info := resultBuffer[core]
		coreInfo = append(coreInfo, systemProcessorPerformanceInfoToCPUTimesStat(core, &info))
	}

	return coreInfo, nil
}

// times has the same parameters and return values as gopsutil's cpu.Times() function
// it actually relies on gopsutil's cpu.Times for non-per core information.
// when percore is true it will attempt to use windows dll's to get the information
// or fall back to gopsutil
func (m *Monitor) times(perCore bool) ([]cpu.TimesStat, error) {
	// non-percore utilization in gopsutil does not rely on wmi so it's fine to
	// utilize it as is
	if !perCore {
		return gopsutilTimes(perCore)
	}
	// Underneath the hood gopsutil relies on a wmi query for per core cpu utilization information
	// this wmi query has proven to be problematic under unclear conditions.  It will hang
	// from time to time, and when executed frequently.  Many projects rely on gopsutil for this information.
	// Some have issues open complaining about hanging wmi calls, but none have a clear solution.
	// In general if you search for information about WMI it is the best of a bad situation.  It
	// is known to be buggy and slow.  A more performant solution is to
	// get the System Processor Performance Information from the ntQuerySystemInformation function.

	// attempt to use the windows api to get information on per core cpu times
	res, err := ntQuerySystemInformation()
	if err == nil {
		return res, nil
	}

	// fall back to gopsutil if there was an error or the dll and proc weren't loaded/found
	m.logger.WithField("debug", err).Debugf("falling back to gopsutil for per core cpu times")
	return gopsutilTimes(perCore)
}
