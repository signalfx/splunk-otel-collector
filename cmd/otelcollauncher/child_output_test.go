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

package main

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForwardOutputStreamTrimsLineEndingsSkipsEmptyLinesAndEmitsFinalPartialLine(t *testing.T) {
	var lines []string

	err := forwardOutputStream(strings.NewReader("one\r\n\r\ntwo\n\nthree"), func(msg string) {
		lines = append(lines, msg)
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, lines)
}

func TestForwardOutputStreamHandlesLongLines(t *testing.T) {
	longLine := strings.Repeat("x", 128*1024)
	var lines []string

	err := forwardOutputStream(strings.NewReader(longLine+"\n"), func(msg string) {
		lines = append(lines, msg)
	})

	require.NoError(t, err)
	assert.Equal(t, []string{longLine}, lines)
}

func TestForwardChildOutputMapsStreamsToSinkLevels(t *testing.T) {
	sink := &recordingSink{}

	err := <-forwardChildOutput(
		strings.NewReader("stdout line\n"),
		strings.NewReader("stderr line\n"),
		sink,
	)

	require.NoError(t, err)
	assert.Equal(t, []string{"stdout line"}, sink.infoLines())
	assert.Equal(t, []string{"stderr line"}, sink.errorLines())
}

type recordingSink struct {
	info   []string
	errors []string
	mu     sync.Mutex
}

func (s *recordingSink) Info(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.info = append(s.info, msg)
}

func (s *recordingSink) Error(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = append(s.errors, msg)
}

func (s *recordingSink) infoLines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.info...)
}

func (s *recordingSink) errorLines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.errors...)
}
