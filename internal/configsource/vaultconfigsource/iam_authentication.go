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
	"errors"
	"fmt"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-secure-stdlib/awsutil"
	"github.com/hashicorp/vault/api"
)

// IAMAuthentication holds the authentication options for AWS IAM. The options
// are the same as the vault CLI tool, see https://github.com/hashicorp/vault/blob/v1.1.0/builtin/credential/aws/cli.go#L148.
type IAMAuthentication struct {
	// AWSAccessKeyID is the AWS access key ID.
	AWSAccessKeyID *string `mapstructure:"aws_access_key_id"`
	// AWSSecretAccessKey is the AWS secret access key.
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

// Token performs an AWS IAM login against the Vault server and returns the
// resulting client token. The logic mirrors upstream's
// builtin/credential/aws/cli.go CLIHandler.Auth, inlined here to avoid
// depending on the github.com/hashicorp/vault module.
func (iam *IAMAuthentication) Token(client *api.Client) (string, error) {
	mount := "aws"
	if iam.Mount != nil {
		mount = *iam.Mount
	}

	role := ""
	if iam.Role != nil {
		role = *iam.Role
	}

	headerValue := ""
	if iam.HeaderValue != nil {
		headerValue = *iam.HeaderValue
	}

	var accessKeyID, secretAccessKey, securityToken string
	if iam.AWSAccessKeyID != nil {
		accessKeyID = *iam.AWSAccessKeyID
	}
	if iam.AWSSecretAccessKey != nil {
		secretAccessKey = *iam.AWSSecretAccessKey
	}
	if iam.AWSSecurityToken != nil {
		securityToken = *iam.AWSSecurityToken
	}

	logger := hclog.Default()
	logger.SetLevel(hclog.Info)
	creds, err := awsutil.RetrieveCreds(accessKeyID, secretAccessKey, securityToken, logger)
	if err != nil {
		return "", err
	}

	// The CLI has always defaulted to "us-east-1" if a region is not provided,
	// matching upstream behavior.
	region := awsutil.DefaultRegion
	loginData, err := awsutil.GenerateLoginData(creds, headerValue, region, logger)
	if err != nil {
		return "", err
	}
	if loginData == nil {
		return "", errors.New("got nil response from GenerateLoginData")
	}
	loginData["role"] = role

	secret, err := client.Logical().Write(fmt.Sprintf("auth/%s/login", mount), loginData)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", errors.New("empty response from credential provider")
	}

	return secret.Auth.ClientToken, nil
}
