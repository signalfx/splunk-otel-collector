// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package supportbundle

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

// processName is the collector binary name, compared case-insensitively and
// without a ".exe" suffix so it matches on both Linux and Windows.
const processName = "otelcol"

const processSampleCount = 3

// Tests shorten these intervals.
var (
	processSampleInterval = 2 * time.Second

	diskCollectionTimeout = 10 * time.Second
)

// collectHostInfo gathers collector process, disk, and memory stats.
func collectHostInfo(ctx context.Context, logger *logger, tempDir string) {
	logger.Info("Getting host information...")

	collectProcessSamples(ctx, logger, tempDir)
	collectDiskUsage(ctx, logger, tempDir)
	collectMemoryInfo(ctx, logger, tempDir)
}

// procHandle allows tests to supply process readings without OS queries.
type procHandle interface {
	Pid() int32
	Name() (string, error)
	Status() ([]string, error)
	MemoryInfo() (*process.MemoryInfoStat, error)
	PercentWithContext(ctx context.Context, interval time.Duration) (float64, error)
	IsRunningWithContext(ctx context.Context) (bool, error)
}

// realProcHandle exposes the Process.Pid field as a method.
type realProcHandle struct {
	*process.Process
}

func (r realProcHandle) Pid() int32 { return r.Process.Pid }

// discoverProcesses is replaceable in tests.
var discoverProcesses = func(ctx context.Context) ([]procHandle, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var found []procHandle
	for _, p := range procs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, err := p.Name()
		if err == nil && strings.TrimSuffix(strings.ToLower(name), ".exe") == processName {
			found = append(found, realProcHandle{p})
		}
	}
	return found, nil
}

type sampledProcess struct {
	handle    procHandle
	cpuPrimed bool
}

func collectProcessSamples(ctx context.Context, logger *logger, tempDir string) {
	if ctx.Err() != nil {
		return
	}
	handles, err := discoverProcesses(ctx)
	if err != nil {
		logger.Warnf("failed to list processes: %v", err)
	}
	if ctx.Err() != nil {
		return
	}
	if len(handles) == 0 {
		logger.Warn("Unable to find otelcol PIDs")
		logger.Warn("      process samples will not be collected")
		return
	}

	dest := filepath.Join(tempDir, metricsDirName, "top.txt")
	f, err := createOutputFile(dest)
	if err != nil {
		logger.Warnf("failed to create %s: %v", dest, err)
		return
	}

	var writeErr error
	defer func() {
		if err := finishOutputFile(f, dest, writeErr); err != nil {
			logger.Warnf("failed to write %s: %v", dest, err)
		}
	}()

	writeErr = runProcessSampling(ctx, logger, f, handles)
}

// runProcessSampling samples only the processes found during discovery.
func runProcessSampling(ctx context.Context, logger *logger, w io.Writer, handles []procHandle) error {
	processes := make([]sampledProcess, len(handles))
	for i, h := range handles {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := h.PercentWithContext(ctx, 0)
		if err != nil {
			logger.Warnf("failed to establish CPU baseline for pid %d: %v", h.Pid(), err)
		}
		processes[i] = sampledProcess{handle: h, cpuPrimed: err == nil}
	}

	for i := 0; i < processSampleCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(processSampleInterval):
		}

		alive := processes[:0]
		for _, p := range processes {
			if err := ctx.Err(); err != nil {
				return err
			}
			running, err := p.handle.IsRunningWithContext(ctx)
			if err != nil || !running {
				continue
			}
			if len(alive) == 0 {
				if _, err := fmt.Fprintf(w, "Sample %d/%d at %s\n\n", i+1, processSampleCount, time.Now().Format(time.RFC3339)); err != nil {
					return err
				}
			}
			if err := writeProcessInfo(ctx, w, &p); err != nil {
				return err
			}
			alive = append(alive, p)
		}
		clear(processes[len(alive):])
		processes = alive
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(processes) == 0 {
			logger.Warn("otelcol process(es) exited; stopping further sampling")
			return nil
		}
	}
	return ctx.Err()
}

func writeProcessInfo(ctx context.Context, w io.Writer, sp *sampledProcess) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p := sp.handle
	name, err := p.Name()
	if err != nil {
		name = "unavailable"
	}

	status := "unavailable"
	if statuses, err := p.Status(); err == nil && len(statuses) > 0 {
		status = strings.Join(statuses, ",")
	}

	cpuStr := "unavailable"
	if v, err := p.PercentWithContext(ctx, 0); err == nil {
		// The first successful read establishes a baseline; only later
		// samples have a full interval of CPU usage to report.
		if sp.cpuPrimed {
			cpuStr = fmt.Sprintf("%.1f", v)
		}
		sp.cpuPrimed = true
	}

	var rss any = "unavailable"
	if info, err := p.MemoryInfo(); err == nil && info != nil {
		rss = info.RSS
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "PID: %d\nName: %q\nStatus: %s\nCPU usage (%%): %s\nResident memory (bytes): %v\n\n", p.Pid(), name, status, cpuStr, rss); err != nil {
		return err
	}
	return ctx.Err()
}

// Disk calls are replaceable in tests.
var (
	listPartitions = func(ctx context.Context) ([]disk.PartitionStat, error) {
		return disk.PartitionsWithContext(ctx, true)
	}
	diskUsage = disk.Usage
)

func collectDiskUsage(ctx context.Context, logger *logger, tempDir string) {
	dest := filepath.Join(tempDir, metricsDirName, "df.txt")
	f, err := createOutputFile(dest)
	if err != nil {
		logger.Warnf("failed to create %s: %v", dest, err)
		return
	}

	var writeErr error
	defer func() {
		if err := finishOutputFile(f, dest, writeErr); err != nil {
			logger.Warnf("failed to write %s: %v", dest, err)
		}
	}()

	writeErr = writeDiskUsage(ctx, logger, f)
}

func writeDiskUsage(ctx context.Context, logger *logger, w io.Writer) error {
	diskCtx, cancel := context.WithTimeout(ctx, diskCollectionTimeout)
	defer cancel()

	partitionsFn, usageFn := listPartitions, diskUsage
	partitions, err := queryDisk(diskCtx, func() ([]disk.PartitionStat, error) {
		return partitionsFn(diskCtx)
	})
	if err != nil {
		return fmt.Errorf("list disk partitions: %w", err)
	}

	if _, err := fmt.Fprint(w, "Disk usage (bytes)\n\n"); err != nil {
		return err
	}

	for _, p := range partitions {
		mountpoint := p.Mountpoint
		usage, err := queryDisk(diskCtx, func() (*disk.UsageStat, error) {
			return usageFn(mountpoint)
		})
		if err != nil {
			// Cancellation fails the report; a mount failure only skips that mount.
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
			if diskCtx.Err() != nil {
				logger.Warnf("disk collection timed out after %s; keeping collected data", diskCollectionTimeout)
				break
			}
			logger.Warnf("skipping disk usage for %s: %v", p.Mountpoint, err)
			continue
		}
		if _, err := fmt.Fprintf(w, "Device: %q\nMountpoint: %q\nTotal: %d\nUsed: %d\nAvailable: %d\nUsed (%%): %.1f\n\n",
			p.Device, mountpoint, usage.Total, usage.Used, usage.Free, usage.UsedPercent); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// Native disk calls may ignore cancellation, so run them separately to allow
// the caller to stop waiting. Collection stops after a timeout.
func queryDisk[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	type result struct {
		value T
		err   error
	}
	resultCh := make(chan result, 1)
	go func() {
		value, err := fn()
		resultCh <- result{value: value, err: err}
	}()

	select {
	case r := <-resultCh:
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		return r.value, r.err
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

func collectMemoryInfo(ctx context.Context, logger *logger, tempDir string) {
	dest := filepath.Join(tempDir, metricsDirName, "free.txt")
	f, err := createOutputFile(dest)
	if err != nil {
		logger.Warnf("failed to create %s: %v", dest, err)
		return
	}

	var writeErr error
	defer func() {
		if err := finishOutputFile(f, dest, writeErr); err != nil {
			logger.Warnf("failed to write %s: %v", dest, err)
		}
	}()

	writeErr = writeMemoryInfo(ctx, f)
}

func writeMemoryInfo(ctx context.Context, w io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("read virtual memory stats: %w", err)
	}

	if _, err = fmt.Fprintf(w, "Physical memory (bytes)\nTotal: %d\nUsed: %d\nFree: %d\nAvailable: %d\n\n", vm.Total, vm.Used, vm.Free, vm.Available); err != nil {
		return err
	}

	sm, err := mem.SwapMemoryWithContext(ctx)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if err != nil {
		_, err = fmt.Fprintf(w, "Swap: unavailable (%v)\n", err)
	} else {
		_, err = fmt.Fprintf(w, "Swap (bytes)\nTotal: %d\nUsed: %d\nFree: %d\n", sm.Total, sm.Used, sm.Free)
	}
	if err != nil {
		return err
	}
	return ctx.Err()
}
