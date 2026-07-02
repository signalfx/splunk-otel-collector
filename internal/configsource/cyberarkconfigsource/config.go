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

package cyberarkconfigsource

import (
	"time"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

// Config holds the configuration for the creation of CyberArk config source objects.
type Config struct {
	configsource.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	// RetrievalMode selects the CyberArk backend used to fetch secrets. Currently only
	// "cp" (Credential Provider, via the local CLIPasswordSDK binary) is implemented.
	// "ccp" (Central Credential Provider REST API) is reserved for a future release.
	// Defaults to "cp".
	RetrievalMode string `mapstructure:"retrieval_mode"`
	// BinaryPath is the path to the CLIPasswordSDK executable used in "cp" mode. It may
	// be an absolute path or a name resolvable on PATH. Defaults to "CLIPasswordSDK".
	BinaryPath string `mapstructure:"binary_path"`
	// AppID is the CyberArk Application ID (AppDescs.AppID) authorized to retrieve the
	// object. Required.
	AppID string `mapstructure:"app_id"`
	// Safe is the CyberArk safe that holds the object. Required.
	Safe string `mapstructure:"safe"`
	// Folder is the folder within the safe. Optional; when empty "Root" is used when
	// building the query.
	Folder string `mapstructure:"folder"`
	// Object is the name of the CyberArk object (account) to retrieve. Required.
	Object string `mapstructure:"object"`
	// AutoRefresh controls whether the config source watches for credential changes. When
	// false (the default) values are fetched once at config resolution and never watched.
	// When true a polling watcher re-fetches on PollInterval and triggers a collector
	// config reload when the retrieved values change.
	AutoRefresh bool `mapstructure:"auto_refresh"`
	// PollInterval is the interval at which the config source re-fetches the object when
	// AutoRefresh is true. Defaults to 1 minute. Ignored when AutoRefresh is false.
	PollInterval time.Duration `mapstructure:"poll_interval"`
}
