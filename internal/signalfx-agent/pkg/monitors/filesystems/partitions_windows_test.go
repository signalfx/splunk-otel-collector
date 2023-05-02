//go:build windows
// +build windows

package filesystems

import (
	"fmt"
	"testing"
	"unsafe"

	gopsutil "github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const uninitialized = 999
const closed = 998
const compressFlag = uint32(16)     // 0x00000010
const readOnlyFlag = uint32(524288) // 0x00080000

type volumeMock struct {
	name      string
	paths     []string
	driveType uint32
	fsType    string
	fsFlags   uint32
	err       error
}

type volumesMock struct {
	handle  int
	volumes []*volumeMock
}

var driveVolume = func() *volumeMock {
	return &volumeMock{
		name:      "\\\\?\\Volume{1e1e1111-0000-0000-0000-010000000000}\\",
		paths:     []string{"C:\\"},
		driveType: windows.DRIVE_FIXED,
		fsType:    "NTFS",
		fsFlags:   compressFlag,
		err:       nil}
}

var driveAndFolderVolume = func() *volumeMock {
	return &volumeMock{
		name:      "\\\\?\\Volume{0000cccc-0000-0000-0000-010000000000}\\",
		paths:     []string{"D:\\", "C:\\mnt\\driveD\\"},
		driveType: windows.DRIVE_FIXED,
		fsType:    "NTFS",
		fsFlags:   compressFlag | readOnlyFlag,
		err:       nil}
}

var removableDriveVolume = func() *volumeMock {
	return &volumeMock{
		name:      "\\\\?\\Volume{bbbbaaaa-0000-0000-0000-010000000000}\\",
		paths:     []string{"A:\\"},
		driveType: windows.DRIVE_REMOVABLE,
		fsType:    "FAT16",
		fsFlags:   compressFlag,
		err:       nil}
}

func TestGetPartitions_Supersets_gopsutil_PartitionStats(t *testing.T) {
	// Partition stats from gopsutil are for drive mounts only.
	want, err := gopsutil.Partitions(true)
	require.NoError(t, err)

	require.NotEmpty(t, want, "cannot find partition stats using gopsutil")

	var got []gopsutil.PartitionStat
	// getPartitions includes partition stats for drive and folder mounts.
	got, err = getPartitions(true)
	require.NoError(t, err)

	require.NotEmpty(t, got, "cannot find partition stats using getPartitions")

	// Asserting `got` getPartitions stats superset of `want` gopsutil stats.
	assert.Subset(t, got, want)
}

func TestGetPartitionsWin(t *testing.T) {
	type wantType struct {
		stats    []gopsutil.PartitionStat
		numStats int
		hasError bool
	}

	tests := []struct {
		name    string
		volumes *volumesMock
		want    wantType
	}{
		{
			name: "should get all partition stats for no errors",
			volumes: func() *volumesMock {
				firstVolume, nextVolume1, nextVolume2 := driveVolume(), driveAndFolderVolume(), removableDriveVolume()
				vols := append(make([]*volumeMock, 0), firstVolume, nextVolume1, nextVolume2)
				return &volumesMock{handle: 0, volumes: vols}
			}(),
			want: wantType{
				numStats: 4,
				hasError: false,
				stats: []gopsutil.PartitionStat{
					{Device: "C:", Mountpoint: "C:", Fstype: "NTFS", Opts: "rw.compress"},
					{Device: "D:", Mountpoint: "D:", Fstype: "NTFS", Opts: "ro.compress"},
					{Device: "C:\\mnt\\driveD", Mountpoint: "C:\\mnt\\driveD", Fstype: "NTFS", Opts: "ro.compress"},
					{Device: "A:", Mountpoint: "A:", Fstype: "FAT16", Opts: "rw.compress"}},
			},
		},
		{
			name: "should get no partition stats if first volume not found",
			volumes: func() *volumesMock {
				firstVolume, nextVolume1, nextVolume2 := driveVolume(), driveAndFolderVolume(), removableDriveVolume()
				firstVolume.err = fmt.Errorf("volume not found")
				vols := append(make([]*volumeMock, 0), firstVolume, nextVolume1, nextVolume2)
				return &volumesMock{handle: 0, volumes: vols}
			}(),
			want: wantType{
				numStats: 0,
				hasError: true,
				stats:    []gopsutil.PartitionStat{},
			},
		},
		{
			name: "should get partition stats for found next volumes",
			volumes: func() *volumesMock {
				firstVolume, nextVolume1, nextVolume2 := driveVolume(), driveAndFolderVolume(), removableDriveVolume()
				nextVolume1.err = fmt.Errorf("volume not found")
				vols := append(make([]*volumeMock, 0), firstVolume, nextVolume1, nextVolume2)
				return &volumesMock{handle: 0, volumes: vols}
			}(),
			want: wantType{
				numStats: 2,
				hasError: true,
				stats: []gopsutil.PartitionStat{
					{Device: "C:", Mountpoint: "C:", Fstype: "NTFS", Opts: "rw.compress"},
					{Device: "A:", Mountpoint: "A:", Fstype: "FAT16", Opts: "rw.compress"}},
			},
		},
	}

	for _, tst := range tests {
		test := tst
		t.Run(test.name, func(t *testing.T) {
			stats, err := getPartitionsWin(
				test.volumes.getDriveTypeMock,
				test.volumes.findFirstVolumeMock,
				test.volumes.findNextVolumeMock,
				test.volumes.findVolumeCloseMock,
				test.volumes.getVolumePathsMock,
				test.volumes.getFsNameAndFlagsMock)

			if !test.want.hasError {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, test.want.numStats, len(stats), "Number of partition stats not equal to expected")

			for i := 0; i < test.want.numStats; i++ {
				assert.Equal(t, test.want.stats[i], stats[i])
			}
		})
	}
}

func (v *volumesMock) getDriveTypeMock(rootPath string) (driveType uint32) {
	for _, volume := range v.volumes {
		for _, path := range volume.paths {
			if path == rootPath {
				return volume.driveType
			}
		}
	}
	return windows.DRIVE_UNKNOWN
}

func (v *volumesMock) findFirstVolumeMock(volumeNamePtr *uint16) (windows.Handle, error) {
	firstVolume := v.volumes[v.handle]
	if firstVolume.err != nil {
		return windows.Handle(unsafe.Pointer(&v.handle)), firstVolume.err
	}

	volumeName, err := windows.UTF16FromString(firstVolume.name)
	if err != nil {
		return windows.Handle(unsafe.Pointer(&v.handle)), err
	}

	start := uintptr(unsafe.Pointer(volumeNamePtr))
	size := unsafe.Sizeof(*volumeNamePtr)
	for i := range volumeName {
		*(*uint16)(unsafe.Pointer(start + size*uintptr(i))) = volumeName[i]
	}

	return windows.Handle(unsafe.Pointer(&v.handle)), nil
}

func (v *volumesMock) findNextVolumeMock(volumeIndexHandle windows.Handle, volumeNamePtr *uint16) error {
	volumeIndex := *(*int)(unsafe.Pointer(volumeIndexHandle))
	if volumeIndex == uninitialized {
		return windows.ERROR_NO_MORE_FILES
	}

	nextVolumeIndex := volumeIndex + 1
	if nextVolumeIndex >= len(v.volumes) {
		return windows.ERROR_NO_MORE_FILES
	}
	*(*int)(unsafe.Pointer(volumeIndexHandle)) = nextVolumeIndex

	nextVolume := v.volumes[nextVolumeIndex]
	if nextVolume.err != nil {
		return nextVolume.err
	}

	volumeName, err := windows.UTF16FromString(nextVolume.name)
	if err != nil {
		return err
	}

	start := uintptr(unsafe.Pointer(volumeNamePtr))
	size := unsafe.Sizeof(*volumeNamePtr)
	for i := range volumeName {
		*(*uint16)(unsafe.Pointer(start + size*uintptr(i))) = volumeName[i]
	}

	return err
}

func (v *volumesMock) findVolumeCloseMock(volumeIndexHandle windows.Handle) error {
	volumeIndex := *(*int)(unsafe.Pointer(volumeIndexHandle))
	if volumeIndex != uninitialized {
		*(*int)(unsafe.Pointer(volumeIndexHandle)) = closed
	}
	return nil
}

func (v *volumesMock) getVolumePathsMock(volNameBuf []uint16) (volumePaths []string, err error) {
	volumeName := windows.UTF16ToString(volNameBuf)
	for _, volume := range v.volumes {
		if volume.name == volumeName {
			volumePaths = append(volumePaths, volume.paths...)
		}
	}
	if len(volumePaths) == 0 {
		err = fmt.Errorf("path not found for volume: %s", volumeName)
	}
	return volumePaths, err
}

func (v *volumesMock) getFsNameAndFlagsMock(rootPath string, fsNameBuf []uint16, fsFlags *uint32) error {
	for _, volume := range v.volumes {
		for _, path := range volume.paths {
			if rootPath == path {
				*fsFlags = volume.fsFlags
				fsName, err := windows.UTF16FromString(volume.fsType)
				if err != nil {
					return err
				}
				copy(fsNameBuf, fsName)
				return nil
			}
		}
	}
	return fmt.Errorf("cannot find volume information for volume path %s", rootPath)
}
