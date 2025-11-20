//go:build windows

package processlist

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessList(t *testing.T) {
	// On Windows all parameters are ignored, pass nil, so the benchmark is re-checked in
	// case of changes in the implementation.
	processList, err := ProcessList(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, processList)

	runningProcesses := []*TopProcess{}
	waitingProcesses := []*TopProcess{}
	unknownStatusProcesses := []*TopProcess{}
	for _, p := range processList {
		switch p.Status {
		case "R":
			runningProcesses = append(runningProcesses, p)
		case "S":
			waitingProcesses = append(waitingProcesses, p)
		default:
			unknownStatusProcesses = append(unknownStatusProcesses, p)
		}
	}
	assert.NotEmpty(t, runningProcesses)
	assert.NotEmpty(t, waitingProcesses)
	assert.Empty(t, unknownStatusProcesses)

	t.Logf("Running processes:")
	for _, p := range runningProcesses {
		t.Logf("%d\t\t%q", p.ProcessID, p.Command)
	}
}

var topProcesses []*TopProcess // A global variable to prevent the compiler from optimizing the benchmark away.
func BenchmarkProcessList(b *testing.B) {
	var tp []*TopProcess
	for i := 0; i < b.N; i++ {
		// On Windows all parameters are ignored, pass nil, so the benchmark is re-checked in
		// case of changes in the implementation.
		processList, err := ProcessList(nil, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		tp = processList
	}
	topProcesses = tp
}
