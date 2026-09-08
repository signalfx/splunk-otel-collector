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

//go:build !aix

package vaultconfigsource

import (
	"fmt"

	gcpauth "github.com/hashicorp/vault-plugin-auth-gcp/plugin"
	"github.com/hashicorp/vault/api"
)

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
