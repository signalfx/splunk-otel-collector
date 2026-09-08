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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/otelcol"
	"golang.org/x/sys/windows/svc"
)

func TestIsParentProcessOtelcolLauncher(t *testing.T) {
	// Save original functions and restore after test
	originalParentProcessNameFn := parentProcessNameFn
	t.Cleanup(func() {
		parentProcessNameFn = originalParentProcessNameFn
	})

	tests := []struct {
		err        error
		name       string
		parentName string
		expected   bool
	}{
		{
			name:       "launcher",
			parentName: "otelcollauncher.exe",
			expected:   true,
		},
		{
			name:       "launcher case insensitive",
			parentName: "OtelColLauncher.EXE",
			expected:   true,
		},
		{
			name:       "other process",
			parentName: "opampsupervisor.exe",
		},
		{
			name: "lookup error",
			err:  errors.New("lookup failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentProcessNameFn = func() (string, error) {
				return tt.parentName, tt.err
			}
			assert.Equal(t, tt.expected, isParentProcessOtelcolLauncher())
		})
	}
}

var svcRunError error // A global variable to prevent the compiler from optimizing the benchmark away.
func BenchmarkSvcRunFail(b *testing.B) {
	var err error
	params := otelcol.CollectorSettings{}
	for i := 0; i < b.N; i++ {
		err = svc.Run("", otelcol.NewSvcHandler(params))
		if err == nil {
			b.Fatal("svc.Run should have failed")
		}
	}
	svcRunError = err
}
