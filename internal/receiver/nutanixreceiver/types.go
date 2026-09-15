// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package nutanixreceiver

import (
	"strconv"
	"strings"
	"unicode"
)

type metricStat struct {
	Name  string
	Value float64
}

type nutanixCluster struct {
	ID        string
	Name      string
	Functions []string
	Stats     []metricStat
}

type nutanixHost struct {
	ID          string
	Name        string
	ClusterID   string
	ClusterName string
	Stats       []metricStat
}

type nutanixStorageContainer struct {
	Encrypted         *bool
	ID                string
	Name              string
	ClusterID         string
	ClusterName       string
	Stats             []metricStat
	ReplicationFactor int
}

type nutanixVM struct {
	GuestTools         nutanixGuestTools
	BootType           string
	Name               string
	ClusterID          string
	HostID             string
	PowerState         string
	ID                 string
	ProtectionPolicyID string
	ProtectionType     string
	DiskBuses          []string
	Stats              []metricStat
	NumSockets         int
	NICCount           int
	NumCoresPerSocket  int
	MemoryBytes        int64
	HasGPU             bool
}

type nutanixGuestTools struct {
	Installed          *bool
	Enabled            *bool
	Reachable          *bool
	VSSSnapshotCapable *bool
}

type nutanixVolumeGroup struct {
	ID            string
	Name          string
	ClusterID     string
	SharingStatus string
	Stats         []metricStat
}

type nutanixDisk struct {
	ID          string
	Serial      string
	ClusterID   string
	ClusterName string
	HostID      string
	HostName    string
	StorageTier string
	Stats       []metricStat
}

type nutanixSubnet struct {
	AdvancedNetworking *bool
	External           *bool
	ID                 string
	Name               string
	SubnetType         string
	ClusterIDs         []string
}

type additionalMetricsRequest struct {
	VMs               []nutanixVM
	DataProtection    bool
	Files             bool
	Microsegmentation bool
	Networking        bool
	Objects           bool
	PrismCentral      bool
}

type additionalSnapshot struct {
	Metrics []additionalMetric
	Errors  []error
}

type additionalMetric struct {
	Attributes  map[string]string
	Name        string
	Description string
	Unit        string
	Stats       []metricStat
	Value       float64
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func boolString(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func positiveIntString(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func cloneAttrs(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(attrs))
	for k, v := range attrs {
		cloned[k] = v
	}
	return cloned
}

func normalizeStatMetricName(value string) string {
	value = strings.TrimSpace(value)
	var result strings.Builder
	result.Grow(len(value))
	runes := []rune(value)
	lastSeparator := false
	for i, current := range runes {
		previous := rune(0)
		if i > 0 {
			previous = runes[i-1]
		}
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}
		if unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous) ||
			(unicode.IsUpper(previous) && unicode.IsLower(next))) {
			result.WriteByte('_')
		}
		if unicode.IsLetter(current) || unicode.IsDigit(current) {
			result.WriteRune(unicode.ToLower(current))
			lastSeparator = false
		} else if result.Len() > 0 && !lastSeparator {
			result.WriteByte('_')
			lastSeparator = true
		}
	}
	name := strings.Trim(result.String(), "_")
	if name == "" {
		return "unknown"
	}
	if unicode.IsDigit(rune(name[0])) {
		return "stat_" + name
	}
	return name
}

func normalizeStatValue(name string, value float64) (string, float64) {
	lowerName := strings.ToLower(name)
	switch {
	case strings.Contains(lowerName, "bytes"):
		return "By", value
	case strings.Contains(lowerName, "iops"):
		return "{operation}/s", value
	case strings.Contains(lowerName, "usec"), strings.Contains(lowerName, "microsecond"):
		return "us", value
	case strings.Contains(lowerName, "msec"), strings.Contains(lowerName, "millisecond"):
		return "ms", value
	case strings.Contains(lowerName, "ppm"):
		return "1", value / 1_000_000
	case strings.Contains(lowerName, "percent"), strings.Contains(lowerName, "percentage"):
		return "1", value / 100
	default:
		// Nutanix does not provide a type or unit for every statistic. Keep
		// those statistics separate by name and leave the unit unspecified
		// rather than assigning a potentially incorrect unit.
		return "", value
	}
}

func normalizeEnumName(value string) string {
	value = strings.TrimPrefix(value, "$")
	return strings.ToLower(value)
}

func isPrismCentralCluster(cluster nutanixCluster) bool {
	for _, function := range cluster.Functions {
		if normalizeEnumName(function) == "prism_central" {
			return true
		}
	}
	return false
}
