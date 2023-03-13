package auth

import (
	"fmt"

	gcpauth "github.com/hashicorp/vault-plugin-auth-gcp/plugin"
	"github.com/hashicorp/vault/api"
)

// GCPConfig is the config for the GCP Auth method in Vault
type GCPConfig struct {
	// Required. The name of the role you're requesting a token for.
	Role *string `yaml:"role"`

	// This is usually provided via the -path flag in the "vault login"
	// command, but it can be specified here as well. If specified here, it
	// takes precedence over the value for -path.  Defaults to `gcp`.
	Mount *string `yaml:"mount"`

	// Explicitly specified GCP credentials in JSON string format (not recommended)
	Credentials *string `yaml:"credentials"`

	// Time until the generated JWT expires in minutes. The given IAM role will
	// have a max_jwt_exp field, the time in minutes that all valid
	// authentication JWTs must expire within (from time of authentication).
	// Defaults to 15 minutes, the default max_jwt_exp for a role. Must be less
	// than an hour.
	JWTExp *int `yaml:"jwt_exp"`

	// Service account to generate a JWT for. Defaults to credentials
	// "client_email" if "credentials" specified and this value is not.
	// The actual credential must have the "iam.serviceAccounts.signJWT"
	// permissions on this service account.
	ServiceAccount *string `yaml:"service_account"`

	// Project for the service account who will be authenticating to Vault.
	// Defaults to the credential's "project_id" (if credentials are specified)."
	Project *string `yaml:"project"`
}

var _ Method = &GCPConfig{}

// Name of the auth method
func (ic *GCPConfig) Name() string {
	return "gcp"
}

// GetToken returns a token if successfully logged in via GCP credentials
func (ic *GCPConfig) GetToken(client *api.Client) (*api.Secret, error) {
	h := gcpauth.CLIHandler{}
	data := map[string]string{}

	// Have to only set these if provided to not confuse the Auth method below
	if ic.Mount != nil {
		data["mount"] = *ic.Mount
	}
	if ic.Role != nil {
		data["role"] = *ic.Role
	}
	if ic.Credentials != nil {
		data["credentials"] = *ic.Credentials
	}
	if ic.JWTExp != nil {
		data["jwt_exp"] = fmt.Sprintf("%d", *ic.JWTExp)
	}
	if ic.ServiceAccount != nil {
		data["service_account"] = *ic.ServiceAccount
	}
	if ic.Project != nil {
		data["project"] = *ic.Project
	}

	return h.Auth(client, data)
}
