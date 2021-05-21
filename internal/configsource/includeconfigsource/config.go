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

package includeconfigsource

import (
	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

// Config holds the configuration for the creation of include config source objects.
type Config struct {
	*configprovider.Settings

	// DeleteFiles is used to instruct the config source to delete the
	// files after its content is read. The default value is 'false'.
	// Set it to 'true' to force the deletion of the file as soon
	// as the config source finished using it.
	DeleteFiles bool `mapstructure:"delete_files"`
	// WatchFiles is used to control if the referenced files should
	// be watched for updates or not. The default value is 'false'.
	// Set it to 'true' to watch the referenced files for changes.
	WatchFiles bool `mapstructure:"watch_files"`
}
