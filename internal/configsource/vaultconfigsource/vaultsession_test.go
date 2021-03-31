package vaultconfigsource

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func Test_SessionForKV(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()

	address := "http://localhost:8200"
	token := "vault_dev_token"

	vault := testutils.NewContainer().WithImage(
		"vault",
	).WithEnv(map[string]string{
		"VAULT_DEV_ROOT_TOKEN_ID": token,
		"VAULT_TOKEN":             token,
		"VAULT_ADDR":              "http://127.0.0.1:8200",
	}).WithExposedPorts("8200:8200").WillWaitForPorts(
		"8200",
	).WithCmd("vault kv put secret/hello foo=world k0=v1").Build()

	require.NoError(t, vault.Start(context.Background()))
	defer func() {
		assert.NoError(t, vault.Stop(context.Background()))
	}()

	cs, err := newConfigSource(address, token, "secret/data/hello")
	require.NoError(t, err)
	require.NotNil(t, cs)

	s, err := cs.NewSession(context.Background())
	require.NoError(t, err)
	require.NotNil(t, s)

	retrieved, err := s.Retrieve(context.Background(), "data.foo", nil)
	require.NoError(t, err)
	require.Equal(t, "world", retrieved.Value().(string))
	require.NoError(t, s.RetrieveEnd(context.Background()))

	var watcherErr error
	var doneCh chan struct{}
	doneCh = make(chan struct{})
	go func() {
		defer close(doneCh)
		watcherErr = retrieved.WatchForUpdate()
	}()

	time.Sleep(10 * time.Second)
	require.NoError(t, s.Close(context.Background()))

	<-doneCh
	require.Equal(t, configsource.ErrSessionClosed, watcherErr)
}

func Test_pollingUpdate(t *testing.T) {
	tests := []struct {
		name    string
		address string
		token   string
	}{
		{
			name:    "polling_updates",
			address: "http://localhost:8200",
			token:   "vault_dev_token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, err := newConfigSource(tt.address, tt.token, "secret/data/hello")
			require.NoError(t, err)
			require.NotNil(t, cs)

			s, err := cs.NewSession(context.Background())
			require.NoError(t, err)
			require.NotNil(t, s)

			// Retrieve key foo
			retrievedFoo, err := s.Retrieve(context.Background(), "data.foo", nil)
			require.NoError(t, err)
			require.Equal(t, "world", retrievedFoo.Value().(string))

			// Retrieve key excited
			retrievedExcited, err := s.Retrieve(context.Background(), "data.excited", nil)
			require.NoError(t, err)
			require.Equal(t, "yes", retrievedExcited.Value().(string))

			// RetrieveEnd
			require.NoError(t, s.RetrieveEnd(context.Background()))

			// Only the first retrieved key provides a working watcher.
			require.Equal(t, configsource.ErrWatcherNotSupported, retrievedExcited.WatchForUpdate())

			// Waiting for updates and triggering new reads.
			for i := 0; i < 3; i++ {
				var watcherErr error
				var doneCh chan struct{}
				doneCh = make(chan struct{})
				go func() {
					defer close(doneCh)
					watcherErr = retrievedFoo.WatchForUpdate()
				}()

				// Wait for update.
				<-doneCh
				require.ErrorIs(t, watcherErr, configsource.ErrValueUpdated)

				// Close current session.
				require.NoError(t, s.Close(context.Background()))

				// Create a new session and repeat the process.
				s, err = cs.NewSession(context.Background())
				require.NoError(t, err)
				require.NotNil(t, s)

				// Retrieve key foo
				retrievedFoo, err = s.Retrieve(context.Background(), "data.foo", nil)
				require.NoError(t, err)
				require.Equal(t, "world", retrievedFoo.Value().(string))
			}

			// Wait for close.
			var watcherErr error
			var doneCh chan struct{}
			doneCh = make(chan struct{})
			go func() {
				defer close(doneCh)
				watcherErr = retrievedFoo.WatchForUpdate()
			}()

			require.NoError(t, s.Close(context.Background()))
			<-doneCh
			require.ErrorIs(t, watcherErr, configsource.ErrSessionClosed)
		})
	}

}
