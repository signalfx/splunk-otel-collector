// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"

	"go.opentelemetry.io/collector/otelcol"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
)

func run(params otelcol.CollectorSettings) error {
	// There shouldn't be any reason to use NO_WINDOWS_SERVICE anymore, but,
	// keeping it as a forcing mechanism or if someone is concerned about
	// the cost of attempting to run as a service before falling back to
	// interactive mode.
	//
	// The BenchmarkSvcRunFail measures the overhead of attempting and failing to
	// run as a service:
	//
	// goos: windows
	// goarch: amd64
	// cpu: Intel(R) Core(TM) i9-10885H CPU @ 2.40GHz
	// BenchmarkSvcRunFail-16           8232412              4369 ns/op
	//
	if value, present := os.LookupEnv("NO_WINDOWS_SERVICE"); present && value != "0" {
		return runInteractive(params)
	}

	// No need to supply service name when startup is invoked through
	// the Service Control Manager directly.
	if err := svc.Run("", otelcol.NewSvcHandler(params)); err != nil {
		errno, ok := err.(syscall.Errno)
		if ok && errno == windows.ERROR_FAILED_SERVICE_CONTROLLER_CONNECT {
			// Per https://learn.microsoft.com/en-us/windows/win32/api/winsvc/nf-winsvc-startservicectrldispatchera#return-value
			// this means that the process is not running as a service, so run interactively.
			return runInteractive(params)
		}

		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}
