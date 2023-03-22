//go:build !windows
// +build !windows

// Package codegen exists purely to manipulate the go:generate build order so
// that collectd templates are generated first before the monitor metadata
// module is generated.
package codegen

//go:generate sh -c "GOOS=`go env GOHOSTOS` go build -o monitorcodegen ../../../cmd/monitorcodegen"
//go:generate ./monitorcodegen
