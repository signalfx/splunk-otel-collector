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
	"fmt"
	"time"

	gcpauth "github.com/hashicorp/vault-plugin-auth-gcp/plugin"
	"github.com/hashicorp/vault/api"
)

// GCPAuthentication holds the authentication options for GCP. The options
// are the same as the vault CLI tool, see https://github.com/hashicorp/vault-plugin-auth-gcp/blob/e1f6784b379d277038ca0661606aa8d23791e392/plugin/cli.go#L120.
type GCPAuthentication struct {
	// Role is the name of the role you're requesting a token for. It is required.
	Role *string `mapstructure:"role"`
	// Mount is the path where the GCP credential method is mounted.
	// This is usually provided via the -path flag in the "vault login"
	// command, but it can be specified here as well. If specified here, it
	// takes precedence over the value for -path.  Defaults to `gcp`.
	Mount *string `mapstructure:"mount"`
	// Credentials can be used to specify GCP credentials in JSON string format (not recommended).
	Credentials *string `mapstructure:"credentials"`
	// JWTExp is the time until the generated JWT expires. The given GCP role will
	// have a max_jwt_exp field, the time in minutes that all valid
	// authentication JWTs must expire within (from time of authentication).
	// Defaults to 15 minutes, the default max_jwt_exp for a role. Must be less
	// than an hour.
	JWTExpiration *time.Duration `mapstructure:"jwt_exp"`
	// ServiceAccount used to generate a JWT for. Defaults to credentials
	// "client_email" if "credentials" specified and this value is not.
	ServiceAccount *string `mapstructure:"service_account"`
	// Project for the service account who will be authenticating to Vault.
	// Defaults to the credential's "project_id" (if credentials are specified)."
	Project *string `mapstructure:"project"`
}

func (gcp *GCPAuthentication) Token(client *api.Client) (string, error) {
	data := map[string]string{}

	if gcp.Mount != nil {
		data["mount"] = *gcp.Mount
	}
	if gcp.Role != nil {
		data["role"] = *gcp.Role
	}
	if gcp.Credentials != nil {
		data["credentials"] = *gcp.Credentials
	}
	if gcp.JWTExpiration != nil {
		data["jwt_exp"] = fmt.Sprintf("%d", *gcp.JWTExpiration)
	}
	if gcp.ServiceAccount != nil {
		data["service_account"] = *gcp.ServiceAccount
	}
	if gcp.Project != nil {
		data["project"] = *gcp.Project
	}

	h := gcpauth.CLIHandler{}
	secret, err := h.Auth(client, data)
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, nil
}
