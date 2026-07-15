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
	"bufio"
	"errors"
	"io"
	"strings"
	"sync"
)

// childLogSink is the platform-specific destination for child process output.
type childLogSink interface {
	Info(msg string)
	Error(msg string)
}

// forwardChildOutput drains both child process output streams and writes complete
// lines to the provided sink. The returned channel closes after both streams
// reach EOF so callers can wait for output to drain after the child exits.
func forwardChildOutput(stdout, stderr io.Reader, sink childLogSink) <-chan error {
	done := make(chan error, 1)
	errs := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		errs <- forwardOutputStream(stdout, sink.Info)
	}()
	go func() {
		defer wg.Done()
		errs <- forwardOutputStream(stderr, sink.Error)
	}()

	go func() {
		defer close(done)
		wg.Wait()
		close(errs)

		var err error
		for streamErr := range errs {
			err = errors.Join(err, streamErr)
		}
		done <- err
	}()

	return done
}

// forwardOutputStream reads newline-delimited child process output without
// bufio.Scanner token limits. It trims only trailing CR/LF bytes and still
// emits a final partial line when the stream closes without a newline.
func forwardOutputStream(output io.Reader, write func(string)) error {
	reader := bufio.NewReader(output)
	for {
		line, err := reader.ReadString('\n')
		if line = strings.TrimRight(line, "\r\n"); line != "" {
			write(line)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
