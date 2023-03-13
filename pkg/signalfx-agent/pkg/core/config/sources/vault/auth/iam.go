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
