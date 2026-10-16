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
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hostInfoDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, metricsDirName), 0o755))
	return dir
}

func TestCollectDiskUsage(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Device: "test-disk", Mountpoint: "/mnt/data"}}, nil
	})
	withTestDiskUsage(t, func(path string) (*disk.UsageStat, error) {
		assert.Equal(t, "/mnt/data", path)
		return &disk.UsageStat{
			Total:       1024 * 1024 * 1024,
			Used:        512 * 1024 * 1024,
			Free:        256 * 1024 * 1024,
			UsedPercent: 100.0 * 512 / (512 + 256),
		}, nil
	})
	tempDir := hostInfoDir(t)

	logger, _ := newTestLogger()
	collectDiskUsage(t.Context(), logger, tempDir)

	data, err := os.ReadFile(filepath.Join(tempDir, metricsDirName, "df.txt"))
	require.NoError(t, err)
	assert.Equal(t, `Disk usage (bytes)

Device: "test-disk"
Mountpoint: "/mnt/data"
Total: 1073741824
Used: 536870912
Available: 268435456
Used (%): 66.7

`, string(data))
}

func TestCollectMemoryInfo(t *testing.T) {
	tempDir := hostInfoDir(t)

	logger, _ := newTestLogger()
	collectMemoryInfo(t.Context(), logger, tempDir)

	data, err := os.ReadFile(filepath.Join(tempDir, metricsDirName, "free.txt"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "Physical memory (bytes)\nTotal:")
	assert.Contains(t, string(data), "\nAvailable:")
	assert.Contains(t, string(data), "Swap")
}

func withDiscoverProcesses(t *testing.T, fn func(ctx context.Context) ([]procHandle, error)) {
	t.Helper()
	orig := discoverProcesses
	discoverProcesses = fn
	t.Cleanup(func() { discoverProcesses = orig })
}

func withProcessSampleInterval(t *testing.T, d time.Duration) {
	t.Helper()
	orig := processSampleInterval
	processSampleInterval = d
	t.Cleanup(func() { processSampleInterval = orig })
}

func TestCollectProcessSamples_NoMatch(t *testing.T) {
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		return nil, nil
	})

	tempDir := hostInfoDir(t)

	logger, buf := newTestLogger()

	collectProcessSamples(t.Context(), logger, tempDir)

	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "top.txt"))
	assert.Contains(t, buf.String(), "Unable to find otelcol PIDs")
	assert.Contains(t, buf.String(), "process samples will not be collected")
}

// fakeProc supplies process readings and lifecycle events for sampling tests.
type fakeProc struct {
	percentHook   func(context.Context, int)
	nameErr       error
	statusErr     error
	memErr        error
	percentErr    error
	runningErr    error
	memInfo       *process.MemoryInfoStat
	name          string
	status        []string
	percents      []float64 // consumed in order across calls; last value repeats once exhausted
	percentErrors []error
	percentN      int
	memCalls      int
	pid           int32
	running       bool
}

func (f *fakeProc) Pid() int32                { return f.pid }
func (f *fakeProc) Name() (string, error)     { return f.name, f.nameErr }
func (f *fakeProc) Status() ([]string, error) { return f.status, f.statusErr }
func (f *fakeProc) MemoryInfo() (*process.MemoryInfoStat, error) {
	f.memCalls++
	return f.memInfo, f.memErr
}

func (f *fakeProc) PercentWithContext(ctx context.Context, _ time.Duration) (float64, error) {
	defer func() { f.percentN++ }()
	if f.percentHook != nil {
		f.percentHook(ctx, f.percentN)
	}
	if f.percentErr != nil {
		return 0, f.percentErr
	}
	if f.percentN < len(f.percentErrors) && f.percentErrors[f.percentN] != nil {
		return 0, f.percentErrors[f.percentN]
	}
	if len(f.percents) == 0 {
		return 0, nil
	}
	return f.percents[min(f.percentN, len(f.percents)-1)], nil
}

func (f *fakeProc) IsRunningWithContext(context.Context) (bool, error) {
	return f.running, f.runningErr
}

func TestWriteProcessInfo_FirstSuccessfulCPUReadIsBaseline(t *testing.T) {
	fp := &fakeProc{pid: 1, name: "otelcol", percents: []float64{99.9}}
	sp := &sampledProcess{handle: fp}
	var buf bytes.Buffer
	require.NoError(t, writeProcessInfo(t.Context(), &buf, sp))
	assert.True(t, sp.cpuPrimed)
	assert.Contains(t, buf.String(), "CPU usage (%): unavailable\n")
	assert.NotContains(t, buf.String(), "99.9")
}

func TestCollectProcessSamples_DiscoversOnce(t *testing.T) {
	withProcessSampleInterval(t, 5*time.Millisecond)
	fp := &fakeProc{pid: 1, name: "otelcol", running: true, memInfo: &process.MemoryInfoStat{}}

	calls := 0
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		calls++
		return []procHandle{fp}, nil
	})

	tempDir := hostInfoDir(t)
	logger, _ := newTestLogger()
	collectProcessSamples(t.Context(), logger, tempDir)

	assert.Equal(t, 1, calls, "a stable collector process must be discovered once, not once per sample")
	data, err := os.ReadFile(filepath.Join(tempDir, metricsDirName, "top.txt"))
	require.NoError(t, err)
	assert.Equal(t, processSampleCount, strings.Count(string(data), "Sample "))
}

func TestCollectProcessSamples_DropsExitedProcessWithoutRediscovery(t *testing.T) {
	withProcessSampleInterval(t, 5*time.Millisecond)
	exiting := &fakeProc{pid: 11111, name: "otelcol", running: true}
	stable := &fakeProc{pid: 22222, name: "otelcol", running: true, percents: []float64{0, 12.5}}
	exiting.percentHook = func(_ context.Context, call int) {
		if call == 1 {
			exiting.running = false
		}
	}

	calls := 0
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		calls++
		return []procHandle{exiting, stable}, nil
	})
	tempDir := hostInfoDir(t)
	logger, _ := newTestLogger()
	collectProcessSamples(t.Context(), logger, tempDir)

	assert.Equal(t, 1, calls, "exits must not trigger discovery of replacement processes")
	data, err := os.ReadFile(filepath.Join(tempDir, metricsDirName, "top.txt"))
	require.NoError(t, err)
	out := string(data)
	assert.Equal(t, processSampleCount, strings.Count(out, "Sample "))
	assert.Equal(t, 1, strings.Count(out, "11111"))
	assert.Equal(t, processSampleCount, strings.Count(out, "22222"))
	assert.Equal(t, processSampleCount, strings.Count(out, "12.5"), "compaction must retain the surviving CPU baseline")
}

func TestRunProcessSampling_AllProcessesGone(t *testing.T) {
	withProcessSampleInterval(t, 5*time.Millisecond)
	for _, tc := range []struct {
		err  error
		name string
	}{
		{name: "exited"},
		{name: "liveness-error", err: errors.New("process no longer accessible")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fp := &fakeProc{pid: 1, name: "otelcol", runningErr: tc.err}
			logger, logBuf := newTestLogger()
			var buf bytes.Buffer
			require.NoError(t, runProcessSampling(t.Context(), logger, &buf, []procHandle{fp}))
			assert.Empty(t, buf.String())
			assert.Contains(t, logBuf.String(), "stopping further sampling")
			assert.Equal(t, 1, fp.percentN, "an exited or replaced process must not be sampled")
		})
	}
}

func TestRunProcessSampling_RetriesFailedCPUBaseline(t *testing.T) {
	withProcessSampleInterval(t, 5*time.Millisecond)
	initialErr := errors.New("CPU times unavailable")
	fp := &fakeProc{
		pid:           1,
		name:          "otelcol",
		running:       true,
		percents:      []float64{0, 99.9, 12.5, 3.0},
		percentErrors: []error{initialErr},
	}
	logger, logBuf := newTestLogger()
	var buf bytes.Buffer
	var reads []time.Time
	fp.percentHook = func(context.Context, int) { reads = append(reads, time.Now()) }
	require.NoError(t, runProcessSampling(t.Context(), logger, &buf, []procHandle{fp}))

	assert.Contains(t, logBuf.String(), initialErr.Error())
	assert.NotContains(t, buf.String(), "99.9", "the first successful retry is a baseline, not a sample")
	assert.Contains(t, buf.String(), "12.5")
	assert.Contains(t, buf.String(), "3.0")
	assert.Contains(t, strings.Split(buf.String(), "Sample 2/")[0], "CPU usage (%): unavailable\n")
	require.Len(t, reads, 1+processSampleCount)
	assert.GreaterOrEqual(t, reads[2].Sub(reads[1]), processSampleInterval)
}

func TestWriteProcessInfo_CPUErrorsPreserveBaseline(t *testing.T) {
	fp := &fakeProc{pid: 1, name: "otelcol", percentErr: errors.New("CPU times unavailable"), percents: []float64{0, 12.5}}
	sp := &sampledProcess{handle: fp, cpuPrimed: true}
	var buf bytes.Buffer
	require.NoError(t, writeProcessInfo(t.Context(), &buf, sp))
	assert.Contains(t, buf.String(), "CPU usage (%): unavailable\n")
	assert.True(t, sp.cpuPrimed)

	fp.percentErr = nil
	buf.Reset()
	require.NoError(t, writeProcessInfo(t.Context(), &buf, sp))
	assert.Contains(t, buf.String(), "CPU usage (%): 12.5\n")
}

func TestCollectProcessSamples_DiscoveryFailure(t *testing.T) {
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		return nil, errors.New("process enumeration failed")
	})
	tempDir := t.TempDir()
	logger, logBuf := newTestLogger()
	collectProcessSamples(t.Context(), logger, tempDir)
	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "top.txt"))
	assert.Contains(t, logBuf.String(), "process enumeration failed")
}

func TestCollectProcessSamples_CanceledDiscovery(t *testing.T) {
	for _, alreadyCanceled := range []bool{true, false} {
		t.Run(fmt.Sprintf("already-canceled=%t", alreadyCanceled), func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if alreadyCanceled {
				cancel()
			}
			called := false
			withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
				called = true
				cancel()
				return nil, context.Canceled
			})
			tempDir := hostInfoDir(t)
			logger, logBuf := newTestLogger()
			collectProcessSamples(ctx, logger, tempDir)
			assert.Equal(t, !alreadyCanceled, called)
			assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "top.txt"))
			assert.NotContains(t, logBuf.String(), "Unable to find otelcol PIDs")
		})
	}
}

func TestRunProcessSampling_IntervalCPU(t *testing.T) {
	withProcessSampleInterval(t, 10*time.Millisecond)

	fp := &fakeProc{
		pid:      4242,
		name:     "otelcol",
		status:   []string{"running"},
		memInfo:  &process.MemoryInfoStat{RSS: 100 * 1024 * 1024},
		running:  true,
		percents: []float64{0, 12.5, 47.0, 3.0}, // index 0 is the discarded priming call
	}
	var buf bytes.Buffer
	logger, _ := newTestLogger()

	var reads []time.Time
	fp.percentHook = func(context.Context, int) { reads = append(reads, time.Now()) }
	err := runProcessSampling(t.Context(), logger, &buf, []procHandle{fp})
	require.NoError(t, err)

	// 1 priming call + processSampleCount real samples.
	assert.Equal(t, 1+processSampleCount, fp.percentN)
	assert.Equal(t, processSampleCount, fp.memCalls)
	for i := 1; i < len(reads); i++ {
		assert.GreaterOrEqual(t, reads[i].Sub(reads[i-1]), processSampleInterval)
	}
	out := buf.String()
	assert.Contains(t, out, "CPU usage (%): 12.5\n")
	assert.Contains(t, out, "CPU usage (%): 47.0\n")
	assert.Contains(t, out, "CPU usage (%): 3.0\n")
	assert.Contains(t, out, "Resident memory (bytes): 104857600\n")
	assert.Contains(t, out, "Status: running\n")
	assert.Contains(t, out, "PID: 4242\n")
	assert.Contains(t, out, "Name: \"otelcol\"\n")
}

func TestWriteProcessInfo_FailedMetricsRenderAsUnknown(t *testing.T) {
	fp := &fakeProc{
		pid:        99999, // simulates an invalid/stale PID: every query fails
		nameErr:    errors.New("no such process"),
		statusErr:  errors.New("no such process"),
		memErr:     errors.New("no such process"),
		percentErr: errors.New("no such process"),
	}
	sp := &sampledProcess{handle: fp, cpuPrimed: true}

	var buf bytes.Buffer
	require.NoError(t, writeProcessInfo(t.Context(), &buf, sp))

	out := buf.String()
	assert.Contains(t, out, "99999")
	assert.Contains(t, out, "Name: \"unavailable\"\n")
	assert.Contains(t, out, "Status: unavailable\n")
	assert.Contains(t, out, "CPU usage (%): unavailable\n")
	assert.Contains(t, out, "Resident memory (bytes): unavailable\n")
	assert.NotContains(t, out, "0.0")
}

func TestRunProcessSampling_RespectsCancellation(t *testing.T) {
	fp := &fakeProc{pid: 1, name: "otelcol", running: true}
	logger, _ := newTestLogger()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	var buf bytes.Buffer
	err := runProcessSampling(ctx, logger, &buf, []procHandle{fp})

	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, fp.percentN)
	assert.Empty(t, buf.String())
}

func TestRunProcessSampling_CanceledAfterBaseline(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fp := &fakeProc{pid: 1, name: "otelcol", running: true}
	fp.percentHook = func(context.Context, int) { cancel() }
	logger, _ := newTestLogger()
	var buf bytes.Buffer
	require.ErrorIs(t, runProcessSampling(ctx, logger, &buf, []procHandle{fp}), context.Canceled)
	assert.Equal(t, 1, fp.percentN)
	assert.Zero(t, fp.memCalls)
	assert.Empty(t, buf.String())
}

func TestWriteProcessInfo_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	fp := &fakeProc{pid: 1, name: "otelcol"}
	var buf bytes.Buffer
	require.ErrorIs(t, writeProcessInfo(ctx, &buf, &sampledProcess{handle: fp}), context.Canceled)
	assert.Zero(t, fp.percentN)
	assert.Zero(t, fp.memCalls)
	assert.Empty(t, buf.String())
}

func TestCollectProcessSamples_CancellationRemovesIncompleteReport(t *testing.T) {
	withProcessSampleInterval(t, 5*time.Millisecond)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fp := &fakeProc{pid: 1, name: "otelcol", running: true}
	fp.percentHook = func(_ context.Context, call int) {
		if call == processSampleCount {
			cancel()
		}
	}
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		return []procHandle{fp}, nil
	})

	tempDir := hostInfoDir(t)

	logger, _ := newTestLogger()
	collectProcessSamples(ctx, logger, tempDir)

	assert.Equal(t, 1+processSampleCount, fp.percentN)
	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "top.txt"))
}

type failingWriter struct {
	failAfter int
	writes    int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes > w.failAfter {
		return 0, errors.New("simulated write failure")
	}
	return len(p), nil
}

type writeFunc func([]byte) (int, error)

func (f writeFunc) Write(p []byte) (int, error) { return f(p) }

func TestRunProcessSampling_CancellationDuringLastWrite(t *testing.T) {
	withProcessSampleInterval(t, time.Millisecond)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	fp := &fakeProc{pid: 1, name: "otelcol", running: true}
	w := writeFunc(func(p []byte) (int, error) {
		if fp.percentN == 1+processSampleCount {
			cancel()
		}
		return len(p), nil
	})
	logger, _ := newTestLogger()
	err := runProcessSampling(ctx, logger, w, []procHandle{fp})
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1+processSampleCount, fp.percentN)
}

func TestRunProcessSampling_WriteFailure(t *testing.T) {
	withProcessSampleInterval(t, time.Millisecond)
	for _, failAfter := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("write-%d", failAfter+1), func(t *testing.T) {
			fp := &fakeProc{pid: 1, name: "otelcol", running: true}
			logger, _ := newTestLogger()
			err := runProcessSampling(t.Context(), logger, &failingWriter{failAfter: failAfter}, []procHandle{fp})
			require.ErrorContains(t, err, "simulated write failure")
		})
	}
}

func TestWriteDiskUsage_WriteFailure(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Mountpoint: "/"}}, nil
	})
	withTestDiskUsage(t, func(string) (*disk.UsageStat, error) { return &disk.UsageStat{}, nil })
	for _, failAfter := range []int{0, 1} {
		t.Run(fmt.Sprintf("write-%d", failAfter+1), func(t *testing.T) {
			logger, _ := newTestLogger()
			err := writeDiskUsage(t.Context(), logger, &failingWriter{failAfter: failAfter})
			require.ErrorContains(t, err, "simulated write failure")
		})
	}
}

func TestWriteMemoryInfo_WriteFailure(t *testing.T) {
	for _, failAfter := range []int{0, 1} {
		t.Run(fmt.Sprintf("write-%d", failAfter+1), func(t *testing.T) {
			err := writeMemoryInfo(t.Context(), &failingWriter{failAfter: failAfter})
			require.ErrorContains(t, err, "simulated write failure")
		})
	}
}

func TestCollectHostInfo_ExistingFileIsPreserved(t *testing.T) {
	withDiscoverProcesses(t, func(context.Context) ([]procHandle, error) {
		return []procHandle{&fakeProc{pid: 1, name: "otelcol", running: true}}, nil
	})
	for _, tc := range []struct {
		collect func(context.Context, *logger, string)
		name    string
	}{
		{name: "top.txt", collect: collectProcessSamples},
		{name: "df.txt", collect: collectDiskUsage},
		{name: "free.txt", collect: collectMemoryInfo},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := hostInfoDir(t)
			dest := filepath.Join(tempDir, metricsDirName, tc.name)
			require.NoError(t, os.WriteFile(dest, []byte("pre-existing"), 0o600))
			logger, logBuf := newTestLogger()
			tc.collect(t.Context(), logger, tempDir)
			assert.Contains(t, logBuf.String(), "failed to create")
			got, err := os.ReadFile(dest)
			require.NoError(t, err)
			assert.Equal(t, "pre-existing", string(got))
		})
	}
}

func TestCollectMemoryInfo_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	tempDir := hostInfoDir(t)
	logger, logBuf := newTestLogger()
	collectMemoryInfo(ctx, logger, tempDir)
	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "free.txt"))
	assert.Contains(t, logBuf.String(), context.Canceled.Error())
}

func withTestPartitions(t *testing.T, fn func(ctx context.Context) ([]disk.PartitionStat, error)) {
	t.Helper()
	orig := listPartitions
	listPartitions = fn
	t.Cleanup(func() { listPartitions = orig })
}

func withTestDiskUsage(t *testing.T, fn func(path string) (*disk.UsageStat, error)) {
	t.Helper()
	orig := diskUsage
	diskUsage = fn
	t.Cleanup(func() { diskUsage = orig })
}

func withDiskCollectionTimeout(t *testing.T, timeout time.Duration) {
	t.Helper()
	orig := diskCollectionTimeout
	diskCollectionTimeout = timeout
	t.Cleanup(func() { diskCollectionTimeout = orig })
}

func TestWriteDiskUsage_PartitionsTimeout(t *testing.T) {
	withDiskCollectionTimeout(t, 20*time.Millisecond)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		<-release
		return nil, nil
	})

	logger, _ := newTestLogger()
	var buf bytes.Buffer
	err := writeDiskUsage(t.Context(), logger, &buf)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Empty(t, buf.String())
}

func TestCollectDiskUsage_TimeoutKeepsPartialReport(t *testing.T) {
	withDiskCollectionTimeout(t, 100*time.Millisecond)
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{
			{Device: "fast-disk", Mountpoint: "/"},
			{Device: "slow-nfs", Mountpoint: "/mnt/slow"},
			{Device: "later-disk", Mountpoint: "/mnt/later"},
		}, nil
	})
	release := make(chan struct{})
	withTestDiskUsage(t, func(path string) (*disk.UsageStat, error) {
		if path == "/mnt/slow" {
			<-release
		}
		return &disk.UsageStat{Total: 1024, Used: 512, Free: 512, UsedPercent: 50}, nil
	})
	t.Cleanup(func() { close(release) })

	tempDir := hostInfoDir(t)
	logger, logBuf := newTestLogger()
	collectDiskUsage(t.Context(), logger, tempDir)
	data, err := os.ReadFile(filepath.Join(tempDir, metricsDirName, "df.txt"))
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(logBuf.String(), "timed out"))
	assert.Contains(t, string(data), "fast-disk")
	assert.NotContains(t, string(data), "slow-nfs")
	assert.NotContains(t, string(data), "later-disk")
}

func TestWriteDiskUsage_CancellationDuringLastQuery(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Device: "only-disk", Mountpoint: "/"}}, nil
	})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	withTestDiskUsage(t, func(string) (*disk.UsageStat, error) {
		cancel()
		return &disk.UsageStat{Total: 1}, nil
	})

	logger, _ := newTestLogger()
	var buf bytes.Buffer
	err := writeDiskUsage(ctx, logger, &buf)
	require.ErrorIs(t, err, context.Canceled)
}

func TestCollectDiskUsage_CancellationRemovesIncompleteReport(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Device: "only-disk", Mountpoint: "/"}}, nil
	})

	tempDir := hostInfoDir(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	withTestDiskUsage(t, func(string) (*disk.UsageStat, error) {
		cancel()
		return &disk.UsageStat{Total: 1}, nil
	})

	logger, _ := newTestLogger()
	collectDiskUsage(ctx, logger, tempDir)
	assert.NoFileExists(t, filepath.Join(tempDir, metricsDirName, "df.txt"))
}

func TestWriteDiskUsage_SkipsFailedMount(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{
			{Device: "bad-disk", Mountpoint: "/bad"},
			{Device: "good-disk", Mountpoint: "/good"},
		}, nil
	})
	var paths []string
	withTestDiskUsage(t, func(path string) (*disk.UsageStat, error) {
		paths = append(paths, path)
		if path == "/bad" {
			return nil, errors.New("permission denied")
		}
		return &disk.UsageStat{Total: 1}, nil
	})

	logger, logBuf := newTestLogger()
	var buf bytes.Buffer
	require.NoError(t, writeDiskUsage(t.Context(), logger, &buf))
	assert.Equal(t, []string{"/bad", "/good"}, paths)
	assert.Contains(t, logBuf.String(), "permission denied")
	assert.Contains(t, buf.String(), "good-disk")
	assert.NotContains(t, buf.String(), "bad-disk")
}

func TestWriteDiskUsage_CancellationDuringLastWrite(t *testing.T) {
	withTestPartitions(t, func(context.Context) ([]disk.PartitionStat, error) {
		return []disk.PartitionStat{{Mountpoint: "/"}}, nil
	})
	withTestDiskUsage(t, func(string) (*disk.UsageStat, error) { return &disk.UsageStat{}, nil })
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	writes := 0
	w := writeFunc(func(p []byte) (int, error) {
		writes++
		if writes == 2 {
			cancel()
		}
		return len(p), nil
	})
	logger, _ := newTestLogger()
	require.ErrorIs(t, writeDiskUsage(ctx, logger, w), context.Canceled)
	assert.Equal(t, 2, writes)
}

func TestQueryDisk_ReturnsResult(t *testing.T) {
	value, err := queryDisk(t.Context(), func() (int, error) { return 42, nil })
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestQueryDisk_TimesOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	value, err := queryDisk(ctx, func() (int, error) {
		<-release
		return 42, nil
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Zero(t, value)
}

func TestQueryDisk_CanceledContextDoesNotStartCall(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	_, err := queryDisk(ctx, func() (int, error) {
		called = true
		return 42, nil
	})
	require.ErrorIs(t, err, context.Canceled)
	assert.False(t, called)
}

func TestQueryDisk_CancellationWinsOverResult(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	value, err := queryDisk(ctx, func() (int, error) {
		cancel()
		return 42, nil
	})
	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, value)
}

func TestQueryDisk_PassesThroughUnderlyingError(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := queryDisk(t.Context(), func() (int, error) { return 0, wantErr })
	require.ErrorIs(t, err, wantErr)
}
