package auth

import "github.com/hashicorp/vault/api"

// Method is something that uses a Vault Auth mechanism to get a Vault
// token that can then be used to get secrets from Vault.
type Method interface {
	Name() string
	GetToken(client *api.Client) (*api.Secret, error)
}
