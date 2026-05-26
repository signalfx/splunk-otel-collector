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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChildProcessExitedWithinDetectsImmediateExit(t *testing.T) {
	processDone := make(chan error, 1)
	processDone <- errors.New("child exited")

	err, exited := (&childProcess{processDone: processDone}).exitedWithin(time.Second)

	require.Error(t, err)
	assert.True(t, exited)
}

func TestChildProcessExitedWithinTimesOut(t *testing.T) {
	processDone := make(chan error)

	err, exited := (&childProcess{processDone: processDone}).exitedWithin(time.Millisecond)

	require.NoError(t, err)
	assert.False(t, exited)
}

func TestChildProcessDrainOutputReturnsForwarderError(t *testing.T) {
	outputDone := make(chan error, 1)
	outputDone <- errors.New("read failed")

	err := (&childProcess{outputDone: outputDone}).drainOutput(time.Second)

	require.ErrorContains(t, err, "read failed")
}

func TestChildProcessDrainOutputTimesOutAndClosesReaders(t *testing.T) {
	outputDone := make(chan error)
	closer := &recordingCloser{}

	err := (&childProcess{
		outputDone:    outputDone,
		outputClosers: []io.Closer{closer},
	}).drainOutput(time.Millisecond)

	require.ErrorContains(t, err, "timed out draining child output")
	assert.True(t, closer.closed)
}

type recordingCloser struct {
	closed bool
}

func (c *recordingCloser) Close() error {
	c.closed = true
	return nil
}
