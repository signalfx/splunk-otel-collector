// Copyright Splunk, Inc.
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

package bundle

import (
	"sort"
)

var (
	// These are extensions that must match corresponding bundle.d/extensions/<NAME>.discovery.yaml.tmpl files.
	// If they are desired for !windows BundledFS inclusion (and a default linux conf.d entry), ensure they are included
	// in Components.Linux. If desired in windows BundledFS, ensure they are included in Components.Windows.
	extensions = []string{
		"docker-observer",
		"host-observer",
		"k8s-observer",
	}
	// These are receivers that must match corresponding bundle.d/receivers/<NAME>.discovery.yaml.tmpl files
	// If they are desired for !windows BundledFS inclusion (and a default linux conf.d entry), ensure they are included
	// in Components.Linux. If desired in windows BundledFS, ensure they are included in Components.Windows.
	receivers = []string{
		"apache",
		"envoy",
		"istio",
		"jmx-cassandra",
		"kafkametrics",
		"mongodb",
		"mysql",
		"nginx",
		"oracledb",
		"postgresql",
		"rabbitmq",
		"redis",
		"sqlserver",
	}

	Components = DiscoComponents{
		Extensions: func() []string {
			sort.Strings(extensions)
			return extensions
		}(),
		Receivers: func() []string {
			sort.Strings(receivers)
			return receivers
		}(),
		Linux: func() map[string]struct{} {
			linux := map[string]struct{}{}
			for _, extension := range extensions {
				linux[extension] = struct{}{}
			}
			for _, receiver := range receivers {
				linux[receiver] = struct{}{}
			}
			return linux
		}(),
		Windows: func() map[string]struct{} {
			windows := map[string]struct{}{
				"apache":        {},
				"envoy":         {},
				"istio":         {},
				"jmx-cassandra": {},
				"kafkametrics":  {},
				"mongodb":       {},
				"mysql":         {},
				"oracledb":      {},
				"postgresql":    {},
				"rabbitmq":      {},
				"redis":         {},
				"sqlserver":     {},
			}
			for _, extension := range extensions {
				windows[extension] = struct{}{}
			}
			return windows
		}(),
	}
)

type DiscoComponents struct {
	Linux      map[string]struct{}
	Windows    map[string]struct{}
	Extensions []string
	Receivers  []string
}
