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

package splunksecretconfigsource

import (
	"time"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

// Config defines splunksecretconfigsource configuration.
type Config struct {
	configsource.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	// Endpoint is the Splunk management URI, e.g. "https://127.0.0.1:8089". It is typically
	// set to ${env:SPLUNK_MANAGEMENT_URI} when running as a Splunk modular input.
	Endpoint string `mapstructure:"endpoint"`

	// SessionKey is the Splunk session key used to authenticate requests to the
	// storage/passwords REST endpoint. It is typically set to ${env:SPLUNK_SESSION_KEY}
	// when running as a Splunk modular input.
	SessionKey string `mapstructure:"session_key"`

	// App scopes the storage/passwords lookup to a specific Splunk app namespace
	// (servicesNS/<user>/<app>/storage/passwords). A <realm>:<name> credential is only
	// guaranteed unique within the app it was created in, so if multiple apps store a
	// credential with the same realm and name, set this to the owning app to disambiguate.
	// Defaults to "-", the wildcard documented for servicesNS meaning "all apps", which
	// preserves lookups across every app visible to the session.
	App string `mapstructure:"app"`

	// User scopes the storage/passwords lookup to a specific Splunk user namespace
	// (servicesNS/<user>/<app>/storage/passwords). A <realm>:<name> credential can also be
	// stored with user-level sharing, so if multiple users store a credential with the same
	// realm and name, set this to the owning user to disambiguate. Defaults to "-", the
	// wildcard documented for servicesNS meaning "all users", which preserves lookups
	// across every user visible to the session.
	User string `mapstructure:"user"`

	// InsecureSkipVerify controls whether the TLS certificate presented by the splunkd
	// management endpoint is verified. Defaults to false: certificate verification is
	// enabled unless explicitly disabled.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`

	// Timeout is the maximum amount of time to wait for a response from splunkd.
	// Defaults to 10s if not specified.
	Timeout time.Duration `mapstructure:"timeout"`
}
