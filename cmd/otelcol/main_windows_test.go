//go:build windows
// +build windows

package main

import (
	"testing"

	"go.opentelemetry.io/collector/otelcol"
	"golang.org/x/sys/windows/svc"
)

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
