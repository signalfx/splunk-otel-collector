// Copyright Splunk Inc. 2025
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestLoggerWritesToStderr(t *testing.T) {
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdoutReader.Close()
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stderrReader.Close()

	stdout, stderr := os.Stdout, os.Stderr
	defer func() {
		os.Stdout, os.Stderr = stdout, stderr
	}()
	os.Stdout, os.Stderr = stdoutWriter, stderrWriter
	logger, err := createLogger()
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("OTLP input stream test")
	os.Stdout, os.Stderr = stdout, stderr
	if stdoutCloseErr := stdoutWriter.Close(); stdoutCloseErr != nil {
		t.Fatal(stdoutCloseErr)
	}
	if stderrCloseErr := stderrWriter.Close(); stderrCloseErr != nil {
		t.Fatal(stderrCloseErr)
	}
	stdoutOutput, err := io.ReadAll(stdoutReader)
	if err != nil {
		t.Fatal(err)
	}
	stderrOutput, err := io.ReadAll(stderrReader)
	if err != nil {
		t.Fatal(err)
	}
	if len(stdoutOutput) != 0 {
		t.Fatalf("logger wrote to stdout: %q", stdoutOutput)
	}
	if !strings.Contains(string(stderrOutput), "OTLP input stream test") {
		t.Fatalf("logger did not write to stderr: %q", stderrOutput)
	}
}
