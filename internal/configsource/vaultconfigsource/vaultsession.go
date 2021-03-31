// Copyright 2020 Splunk, Inc.
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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/internal/configsource"
)

type vaultSession struct {
	logger *zap.Logger

	client       *api.Client
	secret       *api.Secret
	path         string
	pollInterval time.Duration

	watcherFn func() error

	doneCh     chan struct{}
	watchersWG sync.WaitGroup
}

var _ configsource.Session = (*vaultSession)(nil)

func (v *vaultSession) Retrieve(_ context.Context, selector string, _ interface{}) (configsource.Retrieved, error) {
	// By default assume that watcher is not supported. The exception will be the first
	// value read from the vault secret.
	watchForUpdateFn := watcherNotSupported

	if v.secret == nil {
		if err := v.readSecret(); err != nil {
			return nil, err
		}

		// Watcher is only supported for the first value retrieved.
		var err error
		watchForUpdateFn, err = v.buildWatcherFn()
		if err != nil {
			return nil, err
		}
	}

	value := traverseToKey(v.secret.Data, selector)
	if value == nil {
		return nil, fmt.Errorf("no value at path %q for key %q", v.path, selector)
	}

	return newRetrieved(value, watchForUpdateFn), nil
}

func (v *vaultSession) RetrieveEnd(context.Context) error {
	return nil
}

func (v *vaultSession) Close(context.Context) error {
	close(v.doneCh)
	v.watchersWG.Wait()

	// Vault doesn't have a close for its client, close completed.
	return nil
}

func newSession(client *api.Client, path string) (*vaultSession, error) {
	// TODO: pass from factory.
	logger, _ := zap.NewDevelopment()
	return &vaultSession{
		logger:       logger,
		client:       client,
		path:         path,
		pollInterval: 2 * time.Second,
		doneCh:       make(chan struct{}),
	}, nil
}

func (v *vaultSession) readSecret() error {
	secret, err := v.client.Logical().Read(v.path)
	if err != nil {
		return err
	}

	// Invalid path does not return error but a nil secret.
	if secret == nil {
		return fmt.Errorf("no secret found at %q", v.path)
	}

	// Incorrect path for v2 return nil data and warnings.
	if secret.Data == nil {
		return fmt.Errorf("no data at %q warnings: %v", v.path, secret.Warnings)
	}

	v.secret = secret
	return nil
}

func (v *vaultSession) buildWatcherFn() (func() error, error) {
	switch {
	// Dynamic secrets can be either renewable or leased.
	case v.secret.Renewable:
		return v.buildRenewerWatcher()
	// TODO: leased secrets need to periodically
	default:
		// Not a dynamic secret the best that can be done is polling.
		return v.buildPollingWatcher()
	}
}

func (v *vaultSession) buildRenewerWatcher() (func() error, error) {
	renewer, err := v.client.NewRenewer(&api.RenewerInput{
		Secret: v.secret,
	})
	if err != nil {
		return nil, err
	}

	watcherFn := func() error {
		v.watchersWG.Add(1)
		defer v.watchersWG.Done()

		go renewer.Renew()
		defer renewer.Stop()

		for {
			select {
			case <-renewer.RenewCh():
				v.logger.Debug("vault secret renewed", zap.String("path", v.path))
			case err := <-renewer.DoneCh():
				// Renewal stopped, error or now the client needs to re-fetch the configuration.
				if err == nil {
					return configsource.ErrValueUpdated
				}
				return err
			case <-v.doneCh:
				return configsource.ErrSessionClosed
			}
		}
	}

	return watcherFn, nil
}

// buildPollingWatcher builds a WatchFotUpdate function that monitors for changes on
// the v.secret metadata. In principle this could be done for the actual value of
// the retrieved keys. However, checking for metadata keeps this in sync with the
// SignalFx SmartAgent behavior.
func (v *vaultSession) buildPollingWatcher() (func() error, error) {
	// Use the same requirements as SignalFx Smart Agent to build a polling watcher for the secret:
	//
	// This secret is not renewable or on a lease.  If it has a
	// "metadata" field and has "/data/" in the vault path, then it is
	// probably a KV v2 secret.  In that case, we do a poll on the
	// secret's metadata to refresh it and notice if a new version is
	// added to the secret.
	mdValue := v.secret.Data["metadata"]
	if mdValue == nil || !strings.Contains(v.path, "/data/") {
		// TODO: Log reason for no support.
		return watcherNotSupported, nil
	}

	mdMap, ok := mdValue.(map[string]interface{})
	if !ok {
		// TODO: Log reason for no support.
		return watcherNotSupported, nil
	}

	originalVersion := v.extractVersionMetadata(mdMap, "created_time", "version")

	watcherFn := func() error {
		v.watchersWG.Add(1)
		defer v.watchersWG.Done()

		metadataPath := strings.Replace(v.path, "/data/", "/metadata/", 1)
		ticker := time.NewTicker(v.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metadataSecret, err := v.client.Logical().Read(metadataPath)
				if err != nil {
					// Docs are not clear about how to differentiate between temporary and permanent errors
					// here. TODO: Count number of consecutive failures before failing
					return err
				}

				if metadataSecret == nil || metadataSecret.Data == nil {
					return fmt.Errorf("no secret metadata found at %q", metadataPath)
				}

				const timestampKey = "updated_time"
				const versionKey = "current_version"
				latestVersion := v.extractVersionMetadata(metadataSecret.Data, timestampKey, versionKey)
				if latestVersion == nil {
					return fmt.Errorf("secret metadata is not in the expected format for keys %q and %q", timestampKey, versionKey)
				}

				// Per SmartAgent code this is enough to trigger an update but it is also possible to check if the
				// the valued of the retrieved keys was changed. The current criteria may trigger updates even for
				// addition of new keys to the secret.
				if originalVersion.Timestamp != latestVersion.Timestamp || originalVersion.Version != latestVersion.Version {
					return configsource.ErrValueUpdated
				}
			case <-v.doneCh:
				return configsource.ErrSessionClosed
			}
		}
	}

	return watcherFn, nil
}

func (v *vaultSession) extractVersionMetadata(metadataMap map[string]interface{}, timestampKey, versionKey string) *versionMetadata {
	timestamp, ok := metadataMap[timestampKey].(string)
	if !ok {
		// TODO: Log reason for no support.
		return nil
	}

	versionNumber, ok := metadataMap[versionKey].(json.Number)
	if !ok {
		// TODO: Log reason for no support.
		return nil
	}

	versionInt, err := versionNumber.Int64()
	if err != nil {
		// TODO: Log reason for no support.
		return nil
	}

	return &versionMetadata{
		Timestamp: timestamp,
		Version:   versionInt,
	}
}

type versionMetadata struct {
	Timestamp string
	Version   int64
}

func watcherNotSupported() error {
	return configsource.ErrWatcherNotSupported
}

type retrieved struct {
	value            interface{}
	watchForUpdateFn func() error
}

var _ configsource.Retrieved = (*retrieved)(nil)

func (r *retrieved) Value() interface{} {
	return r.value
}

func (r *retrieved) WatchForUpdate() error {
	return r.watchForUpdateFn()
}

func newRetrieved(value interface{}, watchForUpdateFn func() error) *retrieved {
	return &retrieved{
		value,
		watchForUpdateFn,
	}
}
