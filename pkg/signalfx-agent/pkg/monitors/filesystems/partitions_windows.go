//go:build windows
// +build windows

package filesystems

import (
	"fmt"
	"strings"
	"syscall"
	"unicode/utf16"

	gopsutil "github.com/shirou/gopsutil/disk"
	"golang.org/x/sys/windows"
)

const volumeNameBufferLength = uint32(windows.MAX_PATH + 1)
const volumePathBufferLength = volumeNameBufferLength

func getPartitions(all bool) ([]gopsutil.PartitionStat, error) {
	return getPartitionsWin(getDriveType, findFirstVolume, findNextVolume, findVolumeClose, getVolumePaths, getFsNameAndFlags)
}

// getPartitions returns partition stats.
// Similar to https://github.com/shirou/gopsutil/blob/7e4dab436b94d671021647dc5dc12c94f490e46e/disk/disk_windows.go#L71
func getPartitionsWin(
	getDriveType func(rootPath string) (driveType uint32),
	findFirstVolume func(volName *uint16) (findVol windows.Handle, err error),
	findNextVolume func(findVol windows.Handle, volName *uint16) (err error),
	findVolumeClose func(findVol windows.Handle) (err error),
	getVolumePaths func(volNameBuf []uint16) ([]string, error),
	getFsNameAndFlags func(rootPath string, fsNameBuf []uint16, fsFlags *uint32) (err error),
) ([]gopsutil.PartitionStat, error) {

	stats := make([]gopsutil.PartitionStat, 0)
	volNameBuf := make([]uint16, volumeNameBufferLength)

	handle, err := findFirstVolume(&volNameBuf[0])
	if err != nil {
		return stats, fmt.Errorf("cannot find first-volume: %w", err)
	}
	defer findVolumeClose(handle)

	var volPaths []string
	if volPaths, err = getVolumePaths(volNameBuf); err != nil {
		return stats, fmt.Errorf("find paths error for first-volume %s: %w", windows.UTF16ToString(volNameBuf), err)
	}

	var partitionStats []gopsutil.PartitionStat

	if len(volPaths) > 0 {
		partitionStats, err = getPartitionStats(getDriveType(volPaths[0]), volPaths, getFsNameAndFlags)
		if err != nil {
			return stats, fmt.Errorf("cannot find partition-stats for first-volume %s: %w", windows.UTF16ToString(volNameBuf), err)
		}
		stats = append(stats, partitionStats...)
	}

	var lastError error
	for {
		volNameBuf = make([]uint16, volumeNameBufferLength)
		if err = findNextVolume(handle, &volNameBuf[0]); err != nil {
			if errno, ok := err.(syscall.Errno); ok && errno == windows.ERROR_NO_MORE_FILES {
				break
			}
			lastError = fmt.Errorf("find next-volume last-error: %w", err)
			continue
		}

		volPaths, err = getVolumePaths(volNameBuf)
		if err != nil {
			lastError = fmt.Errorf("find paths last-error for volume %s: %w", windows.UTF16ToString(volNameBuf), err)
			continue
		}

		if len(volPaths) > 0 {
			partitionStats, err = getPartitionStats(getDriveType(volPaths[0]), volPaths, getFsNameAndFlags)
			if err != nil {
				lastError = fmt.Errorf("find partition-stats last-error for volume %s:q %w", windows.UTF16ToString(volNameBuf), err)
				continue
			}
			stats = append(stats, partitionStats...)
		}
	}

	return stats, lastError
}

func getDriveType(rootPath string) (driveType uint32) {
	rootPathPtr, _ := windows.UTF16PtrFromString(rootPath)
	return windows.GetDriveType(rootPathPtr)
}

func findFirstVolume(volName *uint16) (findVol windows.Handle, err error) {
	return windows.FindFirstVolume(volName, volumeNameBufferLength)
}

func findNextVolume(findVol windows.Handle, volName *uint16) (err error) {
	return windows.FindNextVolume(findVol, volName, volumeNameBufferLength)
}

func findVolumeClose(findVol windows.Handle) (err error) {
	return windows.FindVolumeClose(findVol)
}

// getVolumePaths returns the path for the given volume name.
func getVolumePaths(volNameBuf []uint16) ([]string, error) {
	volPathsBuf := make([]uint16, volumePathBufferLength)
	returnLen := uint32(0)
	if err := windows.GetVolumePathNamesForVolumeName(&volNameBuf[0], &volPathsBuf[0], volumePathBufferLength, &returnLen); err != nil {
		return nil, err
	}

	return split0(volPathsBuf, int(returnLen)), nil
}

// split0 iterates through s16 upto `end` and slices `s16` into sub-slices separated by the null character (uint16(0)).
// split0 converts the sub-slices between the null characters into strings then returns them in a slice.
func split0(s16 []uint16, end int) []string {
	if end > len(s16) {
		end = len(s16)
	}

	from, ss := 0, make([]string, 0)

	for to := 0; to < end; to++ {
		if s16[to] == 0 {
			if from < to && s16[from] != 0 {
				ss = append(ss, string(utf16.Decode(s16[from:to])))
			}
			from = to + 1
		}
	}

	return ss
}

// getFsNameAndFlags sets inputs fsNameBuf and fsFlags with fetched filesystem name (NTFS, FAT32 etc) and flags.
func getFsNameAndFlags(rootPath string, fsNameBuf []uint16, fsFlags *uint32) error {
	volNameBuf := make([]uint16, 256)
	volSerialNum := uint32(0)
	maxComponentLen := uint32(0)
	rootPathPtr, _ := windows.UTF16PtrFromString(rootPath)

	return windows.GetVolumeInformation(
		rootPathPtr,
		&volNameBuf[0],
		uint32(len(volNameBuf)),
		&volSerialNum,
		&maxComponentLen,
		fsFlags,
		&fsNameBuf[0],
		uint32(len(fsNameBuf)))
}

// getPartitionStats returns partition stats for the given volume paths.
// Similar to https://github.com/shirou/gopsutil/blob/master/disk/disk_windows.go#L72
func getPartitionStats(
	driveType uint32,
	volPaths []string,
	getFsNameAndFlags func(rootPath string, fsNameBuf []uint16, fsFlags *uint32) (err error),
) ([]gopsutil.PartitionStat, error) {

	stats := make([]gopsutil.PartitionStat, 0)

	var lastError error
	for _, volPath := range volPaths {
		if driveType == windows.DRIVE_REMOVABLE || driveType == windows.DRIVE_FIXED || driveType == windows.DRIVE_REMOTE || driveType == windows.DRIVE_CDROM {
			fsFlags, fsNameBuf := uint32(0), make([]uint16, 256)

			if err := getFsNameAndFlags(volPath, fsNameBuf, &fsFlags); err != nil {
				// Similar to gopsutil, we avoid setting the likely device-is-not-ready error
				// which happens if there is no disk in the drive.
				if driveType != windows.DRIVE_CDROM && driveType != windows.DRIVE_REMOVABLE {
					lastError = fmt.Errorf("get volume information last-error: %w", err)
				}
				continue
			}

			opts := "rw"
			if int64(fsFlags)&gopsutil.FileReadOnlyVolume != 0 {
				opts = "ro"
			}
			if int64(fsFlags)&gopsutil.FileFileCompression != 0 {
				opts += ".compress"
			}

			p := strings.TrimRight(volPath, "\\")
			stats = append(stats, gopsutil.PartitionStat{
				Device:     p,
				Mountpoint: p,
				Fstype:     windows.UTF16PtrToString(&fsNameBuf[0]),
				Opts:       opts,
			})
		}
	}

	return stats, lastError
}

func getUsage(hostFSPath string, path string) (*gopsutil.UsageStat, error) {
	return gopsutil.Usage(path)
}
