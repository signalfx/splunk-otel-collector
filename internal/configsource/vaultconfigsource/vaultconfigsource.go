package vaultconfigsource

import (
	"context"

	"github.com/hashicorp/vault/api"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

type vaultConfigSource struct {
	client *api.Client
	path   string
}

var _ configsource.ConfigSource = (*vaultConfigSource)(nil)

func (v *vaultConfigSource) NewSession(context.Context) (configsource.Session, error) {
	return newSession(v.client, v.path)
}

func newConfigSource(address, token, path string) (*vaultConfigSource, error) {
	// Client doesn't connect on creation and can't be closed. Keeping the same instance is
	// fine.
	client, err := api.NewClient(&api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return &vaultConfigSource{
		client: client,
		path:   path,
	}, nil
}
