//go:build !windows

package filesystems

import (
	"path/filepath"

	gopsutil "github.com/shirou/gopsutil/v3/disk"

	"github.com/signalfx/signalfx-agent/pkg/utils/hostfs"
)

func getPartitions(all bool) ([]gopsutil.PartitionStat, error) {
	return gopsutil.PartitionsWithContext(hostfs.Context(), all)
}

// getUsage prepends the hostFSPath to the partition mountpoint. This is needed
// when reading from a mounted filesystem, ex: container, as the latest GoPsutil
// now looks into 1/mountinfo and this file does not contain the relative path
func getUsage(hostFSPath, path string) (*gopsutil.UsageStat, error) {
	if hostFSPath != "" {
		path = filepath.Join(hostFSPath, path)
	}
	return gopsutil.UsageWithContext(hostfs.Context(), path)
}
