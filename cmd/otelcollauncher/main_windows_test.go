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

//go:build windows

package main

import (
	"errors"
	"os/exec"
	"testing"

	"golang.org/x/sys/windows"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForChildWaitsForOutputForwarding(t *testing.T) {
	outputDone := make(chan error)
	waitStarted := make(chan struct{})
	waitReleased := make(chan struct{})

	done := waitForChild(outputDone, func() error {
		close(waitStarted)
		<-waitReleased
		return nil
	})

	select {
	case <-waitStarted:
		t.Fatal("wait called before output forwarding completed")
	default:
	}

	outputDone <- nil

	<-waitStarted
	close(waitReleased)
	result := <-done

	require.NoError(t, result.outputErr)
	require.NoError(t, result.waitErr)
}

func TestWaitForChildReturnsOutputAndWaitErrors(t *testing.T) {
	outputErr := errors.New("read failed")
	waitErr := errors.New("wait failed")
	outputDone := make(chan error, 1)
	outputDone <- outputErr

	result := <-waitForChild(outputDone, func() error {
		return waitErr
	})

	assert.ErrorIs(t, result.outputErr, outputErr)
	assert.ErrorIs(t, result.waitErr, waitErr)
}

func TestWaitForChildWithoutOutputForwardingWaitsImmediately(t *testing.T) {
	waitErr := errors.New("wait failed")

	result := <-waitForChild(nil, func() error {
		return waitErr
	})

	require.NoError(t, result.outputErr)
	assert.ErrorIs(t, result.waitErr, waitErr)
}

func TestIsControlCExitCode(t *testing.T) {
	assert.True(t, isControlCExitCode(int(uint32(windows.STATUS_CONTROL_C_EXIT))))
	assert.False(t, isControlCExitCode(1))
}

func TestIsControlCExitWithEmptyExitError(t *testing.T) {
	assert.False(t, isControlCExit(&exec.ExitError{}))
}
