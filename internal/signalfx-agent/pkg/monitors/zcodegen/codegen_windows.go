// Package codegen exists purely to manipulate the go:generate build order so
// that generated monitor sources are created before the monitor metadata
// module is generated.
package codegen

//go:generate go build -o monitorcodegen.exe ../../../cmd/monitorcodegen
//go:generate ./monitorcodegen.exe
