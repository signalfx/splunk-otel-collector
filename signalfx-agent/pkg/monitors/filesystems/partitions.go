// Copyright  Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package filesystems

import (
	"path/filepath"

	gopsutil "github.com/shirou/gopsutil/disk"
)

func getPartitions(all bool) ([]gopsutil.PartitionStat, error) {
	return gopsutil.Partitions(all)
}

// getUsage prepends the hostFSPath to the partition mountpoint. This is needed
// when reading from a mounted filesystem, ex: container, as the latest GoPsutil
// now looks into 1/mountinfo and this file does not contain the relative path
func getUsage(hostFSPath string, path string) (*gopsutil.UsageStat, error) {
	if hostFSPath != "" {
		path = filepath.Join(hostFSPath, path)
	}
	return gopsutil.Usage(path)
}
