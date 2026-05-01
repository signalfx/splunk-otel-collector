//go:build !windows

// Package codegen exists purely to manipulate the go:generate build order so
// that generated monitor sources are created before the monitor metadata
// module is generated.
package codegen

//go:generate sh -c "GOOS=`go env GOHOSTOS` go build -o monitorcodegen ../../../cmd/monitorcodegen"
//go:generate ./monitorcodegen
