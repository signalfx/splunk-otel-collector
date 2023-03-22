package processlist

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessListLinux(t *testing.T) {
	cache := initOSCache()

	tps, err := ProcessList(&Config{}, cache, nil)
	require.Nil(t, err)
	require.Greater(t, len(tps), 0)

	selfPID := os.Getpid()
	for _, proc := range tps {
		if proc.ProcessID == selfPID {
			require.WithinDuration(t, proc.CreatedTime, time.Now(), 1*time.Hour)
			return
		}
	}

	require.Fail(t, "Did not find self pid in process list")
}
