package vault

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

type customWatcher interface {
	ErrorCh() <-chan error
	ShouldRefetchCh() <-chan struct{}
	Run()
	Stop()
}

type kvMetadata struct {
	Timestamp string
	Version   int64
}

type pollingKVV2Watcher struct {
	vaultPath       string
	pollInterval    time.Duration
	client          *api.Client
	latest          kvMetadata
	shouldRefetchCh chan struct{}
	errorCh         chan error
	ctx             context.Context
	cancel          context.CancelFunc
}

func newPollingKVV2Watcher(vaultPath string, secret *api.Secret, client *api.Client, pollInterval time.Duration) (*pollingKVV2Watcher, error) {
	if !strings.Contains(vaultPath, "/data/") {
		return nil, errors.New("vault path does not look like a KV V2 path (missing /data/)")
	}

	var latest kvMetadata
	if mdMap, ok := secret.Data["metadata"].(map[string]interface{}); ok {
		if createdTime, ok := mdMap["created_time"].(string); ok {
			latest.Timestamp = createdTime
		} else {
			return nil, errors.New("kv v2-like secret is missing created_time")
		}

		if v, ok := mdMap["version"].(json.Number); ok {
			if vInt, err := v.Int64(); err == nil {
				latest.Version = vInt
			} else {
				return nil, fmt.Errorf("vault secret metadata.version field is not an integer: %w", err)
			}
		} else {
			return nil, errors.New("kv v2-like secret is missing the version field")
		}
	} else {
		return nil, errors.New("kv v2-like secret is missing the metadata field or it is not a map[string]interface{}")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &pollingKVV2Watcher{
		vaultPath:    vaultPath,
		pollInterval: pollInterval,
		client:       client,
		latest:       latest,
		// This shouldn't be buffered so that if something isn't receiving from
		// the channel, the watch will block since there is no point to doing
		// it.
		shouldRefetchCh: make(chan struct{}),
		errorCh:         make(chan error),
		cancel:          cancel,
		ctx:             ctx,
	}, nil
}

func (p *pollingKVV2Watcher) Run() {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			newMetadata, err := fetchKVMetadata(p.client, p.vaultPath)
			if err != nil {
				select {
				case p.errorCh <- err:
					break
				case <-p.ctx.Done():
					break
				}
				continue
			}
			if newMetadata.Timestamp != p.latest.Timestamp || newMetadata.Version != p.latest.Version {
				select {
				case p.shouldRefetchCh <- struct{}{}:
					break
				case <-p.ctx.Done():
					break
				}
			}
		}
	}
}

func (p *pollingKVV2Watcher) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *pollingKVV2Watcher) ShouldRefetchCh() <-chan struct{} {
	return p.shouldRefetchCh
}

func (p *pollingKVV2Watcher) ErrorCh() <-chan error {
	return p.errorCh
}

func fetchKVMetadata(client *api.Client, path string) (*kvMetadata, error) {
	metadataSecret, err := client.Logical().Read(strings.Replace(path, "/data/", "/metadata/", 1))
	if err != nil {
		return nil, err
	}

	var latest kvMetadata
	// This is similar to the logic that derives the kvMetadata from the
	// original secret but uses different fields.
	if updatedTime, ok := metadataSecret.Data["updated_time"].(string); ok {
		latest.Timestamp = updatedTime
	} else {
		return nil, errors.New("kv v2 secret metadata is missing updated_time")
	}

	if v, ok := metadataSecret.Data["current_version"].(json.Number); ok {
		if vInt, err := v.Int64(); err == nil {
			latest.Version = vInt
		} else {
			return nil, errors.New("vault secret metadata current_version field is not an integer")
		}
	} else {
		return nil, errors.New("kv secret metadata is missing the current_version field")
	}

	return &latest, nil
}
