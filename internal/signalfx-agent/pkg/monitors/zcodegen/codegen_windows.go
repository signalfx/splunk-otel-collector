// Package codegen exists purely to manipulate the go:generate build order so
// that collectd templates are generated first before the monitor metadata
// module is generated.
package codegen

//go:generate go build -o monitorcodegen.exe ../../../cmd/monitorcodegen
//go:generate ./monitorcodegen.exe
