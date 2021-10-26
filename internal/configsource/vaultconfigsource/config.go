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

package vaultconfigsource

import (
	"time"

	expcfg "go.opentelemetry.io/collector/config/experimental/config"
)

// Config holds the configuration for the creation of Vault config source objects.
type Config struct {
	expcfg.SourceSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	// Authentication defines the authentication method to be used.
	Authentication *Authentication `mapstructure:"auth"`
	// Endpoint is the address of the Vault server, typically it is set via the
	// VAULT_ADDR environment variable for the Vault CLI.
	Endpoint string `mapstructure:"endpoint"`
	// Path is the Vault path where the secret to be retrieved is located.
	Path string `mapstructure:"path"`
	// PollInterval is the interval in which the config source will check for
	// changes on the data on the given Vault path. This is only used for
	// non-dynamic secret stores. Defaults to 1 minute if not specified.
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

// Authentication holds the authentication configuration for Vault config source objects.
type Authentication struct {
	// Token is the token to be used to access the Vault server, typically is set
	// via the VAULT_TOKEN environment variable for the Vault CLI.
	Token *string `mapstructure:"token"`
	// IAMAuthentication holds the authentication options for AWS IAM. The options
	// are the same as the vault CLI tool, see https://github.com/hashicorp/vault/blob/v1.1.0/builtin/credential/aws/cli.go#L148.
	IAMAuthentication *IAMAuthentication `mapstructure:"iam"`
	// GCPAuthentication holds the authentication options for GCP. The options
	// are the same as the vault CLI tool, see https://github.com/hashicorp/vault-plugin-auth-gcp/blob/e1f6784b379d277038ca0661606aa8d23791e392/plugin/cli.go#L120.
	GCPAuthentication *GCPAuthentication `mapstructure:"gcp"`
}

func (*Config) Validate() error {
	return nil
}
