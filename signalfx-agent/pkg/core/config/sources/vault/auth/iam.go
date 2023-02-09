// Copyright  Splunk, Inc.
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

package auth

import (
	"github.com/hashicorp/vault/api"
	aws "github.com/hashicorp/vault/builtin/credential/aws"
)

// IAMConfig is the config for the AWS Auth method in Vault
type IAMConfig struct {
	// Explicit AWS access key ID
	AWSAccessKeyID *string `yaml:"awsAccessKeyId"`
	// Explicit AWS secret access key
	AWSSecretAccessKey *string `yaml:"awsSecretAccessKey"`
	// Explicit AWS security token for temporary credentials
	AWSSecurityToken *string `yaml:"awsSecurityToken"`
	// Value for the x-vault-aws-iam-server-id header in requests
	HeaderValue *string `yaml:"headerValue"`
	// Path where the AWS credential method is mounted. This is usually provided
	// via the -path flag in the "vault login" command, but it can be specified
	// here as well. If specified here, it takes precedence over the value for
	// -path. The default value is "aws".
	Mount *string `yaml:"mount"`
	// Name of the Vault role to request a token against
	Role *string `yaml:"role"`
}

var _ Method = &IAMConfig{}

// Name of the auth method
func (ic *IAMConfig) Name() string {
	return "iam"
}

// GetToken returns a token if successfully logged in via AWS IAM credentials
func (ic *IAMConfig) GetToken(client *api.Client) (*api.Secret, error) {
	h := aws.CLIHandler{}
	data := map[string]string{}

	// Have to only set these if provided to not confuse the Auth method below
	if ic.Mount != nil {
		data["mount"] = *ic.Mount
	}
	if ic.Role != nil {
		data["role"] = *ic.Role
	}
	if ic.AWSAccessKeyID != nil {
		data["aws_access_key_id"] = *ic.AWSAccessKeyID
	}
	if ic.AWSSecretAccessKey != nil {
		data["aws_secret_access_key"] = *ic.AWSSecretAccessKey
	}
	if ic.AWSSecurityToken != nil {
		data["aws_security_token"] = *ic.AWSSecurityToken
	}
	if ic.HeaderValue != nil {
		data["header_value"] = *ic.HeaderValue
	}

	return h.Auth(client, data)
}
