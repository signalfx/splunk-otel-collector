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

package main

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const WindowsOs = "windows"

func TestConvertFilePosNonWindows(t *testing.T) {
	if runtime.GOOS == WindowsOs {
		t.Skip("Skipping for Windows")
	}
	migrator := &Migrator{}
	lines := []string{
		"testdata/otelcollector.log\t00000000000057af\t00000000000808f7\n",
		"testdata/calico_node.log\t000000000000026c\t00000000000808cf\n",
		"testdata/loggen.log 00000000000040a7\t00000000000fd4a6\n",
	}
	readers, _ := migrator.ConvertFilePos(lines)

	var expected []*Reader
	reader := &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T02:34:00.905393894Z stderr F 2021/10/01 02:34:00 main.go:189: Set config to /conf/relay.yaml\n2021-10-01T02:34:00.905426919Z stderr F 2021/10/01 02:34:00 main.go:272: Set ballast to 168 MiB\n2021-10-01T02:34:00.905433052Z stderr F 2021/10/01 02:34:00 main.go:286: Set memory limit to 460 MiB\n2021-10-01T02:34:00.931383015Z stderr F 2021-10-01T02:34:00.931Z\tinfo\tservice/collector.go:303\tStarting otelcol...\t{\"Version\": \"v0.33.1\", \"NumCPU\": 2}\n2021-10-01T02:34:00.931397472Z stderr F 2021-10-01T02:34:00.931Z\tinfo\tservice/collector.go:242\tLoading configuration...\n2021-10-01T02:34:01.092707836Z stderr F 2021-10-01T02:34:01.092Z\tinfo\tservice/collector.go:258\tApplying configuration...\n2021-10-01T02:34:01.10836105Z stderr F 2021-10-01T02:34:01.108Z\tinfo\tbuilder/exporters_builder.go:264\tExporter was built.\t{\"kind\": \"exporter\", \"name\": \"splunk_hec/platformMetrics\"}\n2021-10-01T02:34:01.108516901Z stderr F 2021-10-01T02:34:01.108Z\tinfo\tbuilder/exporters_builder.go:264\tExporter was built.\t{\"k")},
		Offset:      22447,
	}
	expected = append(expected, reader)
	reader = &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T02:33:24.124520321Z stderr F 2021-10-01 02:33:24.124 [INFO][1] ipam_plugin.go 75: migrating from host-local to calico-ipam...\n2021-10-01T02:33:24.150828535Z stderr F 2021-10-01 02:33:24.148 [INFO][1] migrate.go 66: checking host-local IPAM data dir dir existence...\n2021-10-01T02:33:24.151033995Z stderr F 2021-10-01 02:33:24.150 [INFO][1] migrate.go 68: host-local IPAM data dir dir not found; no migration necessary, successfully exiting...\n2021-10-01T02:33:24.151072541Z stderr F 2021-10-01 02:33:24.150 [INFO][1] ipam_plugin.go 105: migration from host-local to calico-ipam complete node=\"ip-172-31-4-69\"")},
		Offset:      620,
	}
	expected = append(expected, reader)
	reader = &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T03:05:57.784826827Z stdout F ---begin---\n2021-10-01T03:05:57.791972887Z stdout F num: 1 | loggen--1-lg7qj | 2021-10-01T03:05:57.784755 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\n2021-10-01T03:05:57.791985544Z stdout F num: 2 | loggen--1-lg7qj | 2021-10-01T03:05:57.788103 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\n2021-10-01T03:05:57.791990024Z stdout F num: 3 | loggen--1-lg7qj | 2021-10-01T03:05:57.788130 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\n2021-10-01T03:05:57.791993964Z stdout F num: 4 | loggen--1-lg7qj | 2021-10-01T03:05:57.788148 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\n2021-10-01T03:05:57.791998059Z stdout F num: 5 | loggen--1-lg7qj | 2021-10-01T03:05:57.788167 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\n2021-10-01T03:05:57.792001861Z stdout F num: 6 | loggen--1-lg7qj | 2021-10-01T03:05:57.788184 | rr")},
		Offset:      16551,
	}
	expected = append(expected, reader)

	for i, r := range readers {
		assert.Equal(t, r.Fingerprint.FirstBytes, expected[i].Fingerprint.FirstBytes)
		assert.Equal(t, r.Offset, expected[i].Offset)
	}
}

func TestConvertFilePosWindows(t *testing.T) {
	if runtime.GOOS != WindowsOs {
		t.Skip("Skipping for Non Windows")
	}
	migrator := &Migrator{}
	lines := []string{
		"testdata/otelcollector.log\t00000000000057af\t00000000000808f7\n",
		"testdata/calico_node.log\t000000000000026c\t00000000000808cf\n",
		"testdata/loggen.log 00000000000040a7\t00000000000fd4a6\n",
	}
	readers, _ := migrator.ConvertFilePos(lines)

	var expected []*Reader
	reader := &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T02:34:00.905393894Z stderr F 2021/10/01 02:34:00 main.go:189: Set config to /conf/relay.yaml\r\n2021-10-01T02:34:00.905426919Z stderr F 2021/10/01 02:34:00 main.go:272: Set ballast to 168 MiB\r\n2021-10-01T02:34:00.905433052Z stderr F 2021/10/01 02:34:00 main.go:286: Set memory limit to 460 MiB\r\n2021-10-01T02:34:00.931383015Z stderr F 2021-10-01T02:34:00.931Z\tinfo\tservice/collector.go:303\tStarting otelcol...\t{\"Version\": \"v0.33.1\", \"NumCPU\": 2}\r\n2021-10-01T02:34:00.931397472Z stderr F 2021-10-01T02:34:00.931Z\tinfo\tservice/collector.go:242\tLoading configuration...\r\n2021-10-01T02:34:01.092707836Z stderr F 2021-10-01T02:34:01.092Z\tinfo\tservice/collector.go:258\tApplying configuration...\r\n2021-10-01T02:34:01.10836105Z stderr F 2021-10-01T02:34:01.108Z\tinfo\tbuilder/exporters_builder.go:264\tExporter was built.\t{\"kind\": \"exporter\", \"name\": \"splunk_hec/platformMetrics\"}\r\n2021-10-01T02:34:01.108516901Z stderr F 2021-10-01T02:34:01.108Z\tinfo\tbuilder/exporters_builder.go:264\tExporter was bui")},
		Offset:      22447,
	}
	expected = append(expected, reader)
	reader = &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T02:33:24.124520321Z stderr F 2021-10-01 02:33:24.124 [INFO][1] ipam_plugin.go 75: migrating from host-local to calico-ipam...\r\n2021-10-01T02:33:24.150828535Z stderr F 2021-10-01 02:33:24.148 [INFO][1] migrate.go 66: checking host-local IPAM data dir dir existence...\r\n2021-10-01T02:33:24.151033995Z stderr F 2021-10-01 02:33:24.150 [INFO][1] migrate.go 68: host-local IPAM data dir dir not found; no migration necessary, successfully exiting...\r\n2021-10-01T02:33:24.151072541Z stderr F 2021-10-01 02:33:24.150 [INFO][1] ipam_plugin.go 105: migration from host-local to calico-ipam complete node=\"ip-172-31-4-69\"")},
		Offset:      620,
	}
	expected = append(expected, reader)
	reader = &Reader{
		Fingerprint: &Fingerprint{FirstBytes: []byte("2021-10-01T03:05:57.784826827Z stdout F ---begin---\r\n2021-10-01T03:05:57.791972887Z stdout F num: 1 | loggen--1-lg7qj | 2021-10-01T03:05:57.784755 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\r\n2021-10-01T03:05:57.791985544Z stdout F num: 2 | loggen--1-lg7qj | 2021-10-01T03:05:57.788103 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\r\n2021-10-01T03:05:57.791990024Z stdout F num: 3 | loggen--1-lg7qj | 2021-10-01T03:05:57.788130 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\r\n2021-10-01T03:05:57.791993964Z stdout F num: 4 | loggen--1-lg7qj | 2021-10-01T03:05:57.788148 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\r\n2021-10-01T03:05:57.791998059Z stdout F num: 5 | loggen--1-lg7qj | 2021-10-01T03:05:57.788167 | rrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr\r\n2021-10-01T03:05:57.792001861Z stdout F num: 6 | loggen--1-lg7qj | 2021-10-01T03:05:57.78818")},
		Offset:      16551,
	}
	expected = append(expected, reader)

	for i, r := range readers {
		assert.Equal(t, r.Fingerprint.FirstBytes, expected[i].Fingerprint.FirstBytes)
		assert.Equal(t, r.Offset, expected[i].Offset)
	}
}
