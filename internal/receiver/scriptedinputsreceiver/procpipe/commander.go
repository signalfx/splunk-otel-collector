// Copyright Copyright Splunk, Inc.
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

package procpipe

import (
	"context"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Start the shell and begin watching the process.
// shell's stdout and stderr are written to a file.
func (c *Commander) Start(ctx context.Context) error {
	c.logger.Info("Starting script", zap.String("script", c.execFilePath))

	stdout, _ := exec.Command("which", "sh").Output()
	shLocation := strings.TrimSpace(string(stdout))
	c.logger.Info("shell process started", zap.Any("shLocation", shLocation))

	c.cmd = exec.CommandContext(ctx, shLocation, c.args...) //nolint:gosec

	// Capture standard output and standard error.
	c.cmd.Stdin = strings.NewReader(scripts[c.execFilePath])
	c.cmd.Stdout = c.stdout
	c.cmd.Stderr = c.stdout

	c.doneCh = make(chan struct{}, 1)
	c.waitCh = make(chan struct{})

	if err := c.cmd.Start(); err != nil {
		return err
	}

	c.logger.Debug("shell process started", zap.Any("PID", c.cmd.Process.Pid))
	atomic.StoreInt64(&c.running, 1)

	go c.watch()

	return nil
}

func (c *Commander) Restart(ctx context.Context) error {
	if err := c.Stop(ctx); err != nil {
		return err
	}
	return c.Start(ctx)
}

func (c *Commander) watch() {
	err := c.cmd.Wait()
	if err != nil {
		return
	}
	c.doneCh <- struct{}{}
	atomic.StoreInt64(&c.running, 0)
	close(c.waitCh)
}

// Done returns a channel that will send a signal when the shell process is finished.
func (c *Commander) Done() <-chan struct{} {
	return c.doneCh
}

// Pid returns shell process PID if it is started or 0 if it is not.
func (c *Commander) Pid() int {
	if c.cmd == nil || c.cmd.Process == nil {
		return 0
	}
	return c.cmd.Process.Pid
}

// ExitCode returns shell process exit code if it exited or 0 if it is not.
func (c *Commander) ExitCode() int {
	if c.cmd == nil || c.cmd.ProcessState == nil {
		return 0
	}
	return c.cmd.ProcessState.ExitCode()
}

func (c *Commander) IsRunning() bool {
	return atomic.LoadInt64(&c.running) != 0
}

// Stop the shell process. Sends SIGTERM to the process and wait for up 10 seconds
// and if the process does not finish kills it forcedly by sending SIGKILL.
// Returns after the process is terminated.
func (c *Commander) Stop(ctx context.Context) error {
	if c.cmd == nil || c.cmd.Process == nil {
		// Not started, nothing to do.
		return nil
	}

	// c.logger.Debugf("Stopping shell process, PID=%v", c.cmd.Process.Pid)

	// Gracefully signal process to stop.
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return err
	}

	finished := make(chan struct{})

	// Setup a goroutine to wait a while for process to finish and send kill signal
	// to the process if it doesn't finish.
	var innerErr error
	go func() {
		// Wait 10 seconds.
		t := time.After(10 * time.Second)
		select {
		case <-ctx.Done():
			break
		case <-t:
			break
		case <-finished:
			// Process is successfully finished.
			// c.logger.Debugf("shell process PID=%v successfully stopped.", c.cmd.Process.Pid)
			return
		}

		// Time is out. Kill the process.
		// c.logger.Debugf(
		//	"shell process PID=%d is not responding to SIGTERM. Sending SIGKILL to kill forcedly.",
		//	c.cmd.Process.Pid,
		//)
		if innerErr = c.cmd.Process.Signal(syscall.SIGKILL); innerErr != nil {
			return
		}
	}()

	// Wait for process to terminate
	<-c.waitCh

	atomic.StoreInt64(&c.running, 0)

	// Let goroutine know process is finished.
	close(finished)

	return innerErr
}
