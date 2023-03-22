package filesystems

import (
	"testing"

	gopsutil "github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCommonDimensions(t *testing.T) {
	cases := []struct {
		hostFSPath   string
		ps           *gopsutil.PartitionStat
		expectedDims map[string]string
	}{
		{
			hostFSPath: "/hostfs",
			ps: &gopsutil.PartitionStat{
				Device:     "/dev/sdb1",
				Mountpoint: "/hostfs/var/lib",
				Fstype:     "ext4",
				Opts:       "rw,relatime",
			},
			expectedDims: map[string]string{
				"mountpoint": "/var/lib",
				"device":     "/dev/sdb1",
				"fs_type":    "ext4",
				"mode":       "rw",
			},
		},
		{
			hostFSPath: "/hostfs",
			ps: &gopsutil.PartitionStat{
				Device:     "/dev/sdb1",
				Mountpoint: "/hostfs",
				Fstype:     "ext4",
				Opts:       "ro,relatime",
			},
			expectedDims: map[string]string{
				"mountpoint": "/",
				"device":     "/dev/sdb1",
				"fs_type":    "ext4",
				"mode":       "ro",
			},
		},
		{
			hostFSPath: "",
			ps: &gopsutil.PartitionStat{
				Device:     "/dev/sdb1",
				Mountpoint: "/",
				Fstype:     "ext4",
				Opts:       "rx,relatime",
			},
			expectedDims: map[string]string{
				"mountpoint": "/",
				"device":     "/dev/sdb1",
				"fs_type":    "ext4",
			},
		},
	}

	logger := log.WithFields(log.Fields{"monitorType": monitorType})
	for _, tt := range cases {
		m := Monitor{
			hostFSPath:        tt.hostFSPath,
			sendModeDimension: true,
			logger:            logger,
		}

		dims := m.getCommonDimensions(tt.ps)

		assert.Equal(t, tt.expectedDims, dims)
	}
}
