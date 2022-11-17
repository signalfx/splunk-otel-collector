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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configprovider"
)

const (
	address        = "http://localhost:8200"
	token          = "dev_token"
	vaultContainer = "vault_tests"
	mongoContainer = "mongodb"

	// For the commands below whitespace is used to break the parameters into a correct slice of arguments.

	startVault = "docker run --rm -d -p 8200:8200 -e VAULT_DEV_ROOT_TOKEN_ID=" + token + " -e VAULT_TOKEN=" + token + " -e VAULT_ADDR=http://localhost:8200 --name=" + vaultContainer + " vault"
	stopVault  = "docker stop " + vaultContainer

	setupKVStore  = "docker exec " + vaultContainer + " vault kv put secret/kv k0=v0 k1=v1"
	updateKVStore = "docker exec " + vaultContainer + " vault kv put secret/kv k0=v0 k1=v1.1"

	startMongo            = "docker run --rm -d -p 27017:27017 --name=" + mongoContainer + " mongo"
	stopMongo             = "docker stop " + mongoContainer
	setupDatabaseStore    = "docker exec " + vaultContainer + " vault secrets enable database"
	setupMongoVaultPlugin = "docker exec " + vaultContainer + " vault write database/config/my-mongodb-database plugin_name=mongodb-database-plugin allowed_roles=my-role connection_url=mongodb://host.docker.internal:27017/admin username=\"admin\" password=\"\""
	setupMongoSecret      = "docker exec " + vaultContainer + " vault write database/roles/my-role db_name=my-mongodb-database creation_statements={\"db\":\"admin\",\"roles\":[{\"role\":\"readWrite\"},{\"role\":\"read\",\"db\":\"foo\"}]} default_ttl=2s max_ttl=6s"

	createKVVer1Store = "docker exec " + vaultContainer + " vault secrets enable -version=1 kv"
	setupKVVer1Store  = "docker exec " + vaultContainer + " vault kv put kv/my-secret ttl=8s my-value=s3cr3t"
	setupKVVer1NoTTL  = "docker exec " + vaultContainer + " vault kv put kv/my-secret ttl=0s my-value=s3cr3t"
)

var tokenStr = token // As a variable so it can be used as a pointer.

func TestVaultSessionForKV(t *testing.T) {
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, setupKVStore)

	config := Config{
		Endpoint: address,
		Authentication: &Authentication{
			Token: &tokenStr,
		},
		Path:         "secret/data/kv",
		PollInterval: 2 * time.Second,
	}

	source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	retrieved, err := source.Retrieve(context.Background(), "data.k0", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	val, err := retrieved.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "v0", val)

	retrievedMetadata, err := source.Retrieve(context.Background(), "metadata.version", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called because it is the second retrieve")
	})
	require.NoError(t, err)
	valMetadata, err := retrievedMetadata.AsRaw()
	require.NoError(t, err)
	require.NotNil(t, valMetadata)

	require.NoError(t, retrieved.Close(context.Background()))
	require.NoError(t, retrievedMetadata.Close(context.Background()))

	require.NoError(t, source.Shutdown(context.Background()))
}

func TestVaultPollingKVUpdate(t *testing.T) {
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, setupKVStore)

	config := Config{
		Endpoint: address,
		Authentication: &Authentication{
			Token: &tokenStr,
		},
		Path:         "secret/data/kv",
		PollInterval: 2 * time.Second,
	}

	source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve key "k0"
	watchCh := make(chan *confmap.ChangeEvent, 1)
	retrievedK0, err := source.Retrieve(context.Background(), "data.k0", nil, func(event *confmap.ChangeEvent) {
		watchCh <- event
	})
	require.NoError(t, err)
	valK0, err := retrievedK0.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "v0", valK0)

	// Retrieve key "k1"
	retrievedK1, err := source.Retrieve(context.Background(), "data.k1", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called because it is the second retrieve")
	})
	require.NoError(t, err)
	valK1, err := retrievedK1.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "v1", valK1)

	requireCmdRun(t, updateKVStore)

	// Wait for update.
	ce := <-watchCh
	require.NoError(t, ce.Error)

	// Close current source.
	require.NoError(t, retrievedK0.Close(context.Background()))
	require.NoError(t, retrievedK1.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))

	// Create a new source and repeat the process.
	source, err = newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve key
	retrievedUpdatedK1, err := source.Retrieve(context.Background(), "data.k1", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	valUpdatedK1, err := retrievedUpdatedK1.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "v1.1", valUpdatedK1)

	require.NoError(t, retrievedUpdatedK1.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))
}

func TestVaultRenewableSecret(t *testing.T) {
	// This test is based on the commands described at https://www.vaultproject.io/docs/secrets/databases/mongodb
	requireCmdRun(t, startMongo)
	defer requireCmdRun(t, stopMongo)
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, setupDatabaseStore)
	requireCmdRun(t, setupMongoVaultPlugin)
	requireCmdRun(t, setupMongoSecret)

	config := Config{
		Endpoint: address,
		Authentication: &Authentication{
			Token: &tokenStr,
		},
		Path:         "database/creds/my-role",
		PollInterval: 2 * time.Second,
	}

	source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve key username, it is generated by vault no expected value.
	watchCh := make(chan *confmap.ChangeEvent, 1)
	retrievedUser, err := source.Retrieve(context.Background(), "username", nil, func(event *confmap.ChangeEvent) {
		watchCh <- event
	})
	require.NoError(t, err)
	retrievedUserValue, err := retrievedUser.AsRaw()
	require.NoError(t, err)

	// Retrieve key password, it is generated by vault no expected value.
	retrievedPwd, err := source.Retrieve(context.Background(), "password", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called because it is the second retrieve")
	})
	require.NoError(t, err)
	retrievedPwdValue, err := retrievedPwd.AsRaw()
	require.NoError(t, err)

	ce := <-watchCh
	require.NoError(t, ce.Error)

	// Close current source.
	require.NoError(t, retrievedUser.Close(context.Background()))
	require.NoError(t, retrievedPwd.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))

	// Create a new source and repeat the process.
	source, err = newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve key username, it is generated by vault no expected value.
	retrievedUpdatedUser, err := source.Retrieve(context.Background(), "username", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	retrievedUpdatedUserValue, err := retrievedUpdatedUser.AsRaw()
	require.NoError(t, err)
	require.NotEqual(t, retrievedUserValue, retrievedUpdatedUserValue)

	// Retrieve password and check that it changed.
	retrievedUpdatedPwd, err := source.Retrieve(context.Background(), "password", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called because it is the second retrieve")
	})
	require.NoError(t, err)
	retrievedUpdatedPwdValue, err := retrievedUpdatedPwd.AsRaw()
	require.NoError(t, err)
	require.NotEqual(t, retrievedPwdValue, retrievedUpdatedPwdValue)

	require.NoError(t, retrievedUpdatedUser.Close(context.Background()))
	require.NoError(t, retrievedUpdatedPwd.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))
}

func TestVaultV1SecretWithTTL(t *testing.T) {
	// This test is based on the commands described at https://www.vaultproject.io/docs/secrets/kv/kv-v1
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, createKVVer1Store)
	requireCmdRun(t, setupKVVer1Store)

	config := Config{
		Endpoint: address,
		Authentication: &Authentication{
			Token: &tokenStr,
		},
		Path:         "kv/my-secret",
		PollInterval: 2 * time.Second,
	}

	source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve value
	watchCh := make(chan *confmap.ChangeEvent, 1)
	retrievedValue, err := source.Retrieve(context.Background(), "my-value", nil, func(event *confmap.ChangeEvent) {
		watchCh <- event
	})
	require.NoError(t, err)
	retrievedVal, err := retrievedValue.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", retrievedVal)

	ce := <-watchCh

	// Wait for update.
	require.NoError(t, ce.Error)

	// Close current source.
	require.NoError(t, retrievedValue.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))

	// Create a new source and repeat the process.
	source, err = newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve value
	retrievedValue, err = source.Retrieve(context.Background(), "my-value", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	retrievedVal, err = retrievedValue.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", retrievedVal)

	require.NoError(t, retrievedValue.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))
}

func TestVaultV1NonWatchableSecret(t *testing.T) {
	// This test is based on the commands described at https://www.vaultproject.io/docs/secrets/kv/kv-v1
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, createKVVer1Store)
	requireCmdRun(t, setupKVVer1NoTTL)

	config := Config{
		Endpoint: address,
		Authentication: &Authentication{
			Token: &tokenStr,
		},
		Path:         "kv/my-secret",
		PollInterval: 2 * time.Second,
	}

	source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
	require.NoError(t, err)
	require.NotNil(t, source)

	// Retrieve value
	retrievedValue, err := source.Retrieve(context.Background(), "my-value", nil, func(event *confmap.ChangeEvent) {
		panic("must not be called")
	})
	require.NoError(t, err)
	retrievedVal, err := retrievedValue.AsRaw()
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", retrievedVal)

	// Close current source.
	require.NoError(t, retrievedValue.Close(context.Background()))
	require.NoError(t, source.Shutdown(context.Background()))
}

func TestVaultRetrieveErrors(t *testing.T) {
	requireCmdRun(t, startVault)
	defer requireCmdRun(t, stopVault)
	requireCmdRun(t, setupKVStore)

	ctx := context.Background()

	tests := []struct {
		err      error
		name     string
		path     string
		token    string
		selector string
	}{
		{
			name:  "bad_token",
			path:  "secret/data/kv",
			token: "bad_test_token",
			err:   &errClientRead{},
		},
		{
			name: "non_existent_path",
			path: "made_up_path/data/kv",
			err:  &errNilSecret{},
		},
		{
			name: "v2_missing_data_on_path",
			path: "secret/kv",
			err:  &errNilSecretData{},
		},
		{
			name:     "bad_selector",
			path:     "secret/data/kv",
			selector: "data.missing",
			err:      &errBadSelector{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testToken := token
			if tt.token != "" {
				testToken = tt.token
			}

			config := Config{
				Endpoint: address,
				Authentication: &Authentication{
					Token: &testToken,
				},
				Path:         tt.path,
				PollInterval: 2 * time.Second,
			}

			source, err := newConfigSource(configprovider.CreateParams{Logger: zap.NewNop()}, &config)
			require.NoError(t, err)
			require.NotNil(t, source)

			r, err := source.Retrieve(ctx, tt.selector, nil, nil)
			require.Error(t, err)
			assert.IsType(t, tt.err, err)
			assert.Nil(t, r)
			assert.NoError(t, source.Shutdown(ctx))
		})
	}
}

func Test_vaultSession_extractVersionMetadata(t *testing.T) {
	tests := []struct {
		metadataMap map[string]any
		expectedMd  *versionMetadata
		name        string
	}{
		{
			name: "typical",
			metadataMap: map[string]any{
				"tsKey":  "2021-04-02T22:30:51.4733477Z",
				"verKey": json.Number("1"),
			},
			expectedMd: &versionMetadata{
				Timestamp: "2021-04-02T22:30:51.4733477Z",
				Version:   1,
			},
		},
		{
			name: "missing_expected_timestamp",
			metadataMap: map[string]any{
				"otherKey": "2021-04-02T22:30:51.4733477Z",
				"verKey":   json.Number("1"),
			},
		},
		{
			name: "missing_expected_version",
			metadataMap: map[string]any{
				"tsKey":    "2021-04-02T22:30:51.4733477Z",
				"otherKey": json.Number("1"),
			},
		},
		{
			name: "incorrect_version_format",
			metadataMap: map[string]any{
				"tsKey":  "2021-04-02T22:30:51.4733477Z",
				"verKey": json.Number("not_a_number"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &vaultConfigSource{
				logger: zap.NewNop(),
			}

			metadata := v.extractVersionMetadata(tt.metadataMap, "tsKey", "verKey")
			assert.Equal(t, tt.expectedMd, metadata)
		})
	}
}

func requireCmdRun(t *testing.T, cli string) {
	skipCheck(t)
	parts := strings.Split(cli, " ")
	cmd := exec.Command(parts[0], parts[1:]...) // #nosec
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	time.Sleep(500 * time.Millisecond)
	if err != nil {
		err = fmt.Errorf("cmd.Run() %s %v failed %w. stdout: %q stderr: %q", cmd.Path, cmd.Args, err, stdout.String(), stderr.String())
	}
	require.NoError(t, err)
}

func skipCheck(t *testing.T) {
	if s, ok := os.LookupEnv("RUN_VAULT_DOCKER_TESTS"); ok && s != "" {
		return
	}
	t.Skipf("Test must be explicitly enabled via 'RUN_VAULT_DOCKER_TESTS' environment variable.")
}
