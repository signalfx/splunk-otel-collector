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
	"github.com/hashicorp/vault/api"
	aws "github.com/hashicorp/vault/builtin/credential/aws"
)

// IAMAuthentication holds the authentication options for AWS IAM. The options
// are the same as the vault CLI tool, see https://github.com/hashicorp/vault/blob/v1.1.0/builtin/credential/aws/cli.go#L148.
type IAMAuthentication struct {
	// AWSAccessKeyID is the AWS access key ID.
	AWSAccessKeyID *string `mapstructure:"aws_access_key_id"`
	// AWSSecretAccessKey it the AWS secret access key.
	AWSSecretAccessKey *string `mapstructure:"aws_secret_access_key"`
	// AWSSecurityToken is the AWS security token for temporary credentials.
	AWSSecurityToken *string `mapstructure:"aws_security_token"`
	// HeaderValue for the x-vault-aws-iam-server-id header in requests.
	HeaderValue *string `mapstructure:"header_value"`
	// Mount is the path where the AWS credential method is mounted. The default value is "aws".
	Mount *string `mapstructure:"mount"`
	// Role is the name of the Vault role to request a token against.
	Role *string `mapstructure:"role"`
}

func (iam *IAMAuthentication) Token(client *api.Client) (string, error) {
	data := map[string]string{}

	// Have to only set these if provided to not confuse the Auth method below
	if iam.Mount != nil {
		data["mount"] = *iam.Mount
	}
	if iam.Role != nil {
		data["role"] = *iam.Role
	}
	if iam.AWSAccessKeyID != nil {
		data["aws_access_key_id"] = *iam.AWSAccessKeyID
	}
	if iam.AWSSecretAccessKey != nil {
		data["aws_secret_access_key"] = *iam.AWSSecretAccessKey
	}
	if iam.AWSSecurityToken != nil {
		data["aws_security_token"] = *iam.AWSSecurityToken
	}
	if iam.HeaderValue != nil {
		data["header_value"] = *iam.HeaderValue
	}

	h := aws.CLIHandler{}
	secret, err := h.Auth(client, data)
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, nil
}
