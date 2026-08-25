// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package diskqueuestorageextension

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/xextension/storage"
	"go.uber.org/zap"
)

// special keys used by the persistent queue to store metadata.
const (
	metadataKey = "qmv0"

	// all legacy keys - ignored.
	legacyReadIndexKey                = "ri"
	legacyWriteIndexKey               = "wi"
	legacyCurrentlyDispatchedItemsKey = "di"

	separator = "\n\n"
)

var _ storage.Extension = (*diskQueueStorageExtension)(nil)

func newDiskQueueStorageExtension(settings extension.Settings, cfg *Config) extension.Extension {
	return &diskQueueStorageExtension{
		config:   cfg,
		settings: settings,
	}
}

type diskQueueStorageExtension struct {
	config   *Config
	settings extension.Settings
}

type client struct {
	queue                 *diskQueue
	logger                *zap.Logger
	callbacks             map[string]func()
	metadataFile          *os.File
	name                  string
	path                  string
	metadataWrites        int
	metadataTruncateEvery int
	checkFirstGet         sync.Once
}

func (c *client) Get(_ context.Context, key string) ([]byte, error) {
	// This is a kludgy way to detect that the extension is not used for persistent queueing.
	c.checkFirstGet.Do(func() {
		if key != metadataKey {
			panic("The disk_queue_storage extension can only be used with the persistent queue.")
		}
	})
	switch key {
	case metadataKey:
		b, err := c.readMetadata()
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			c.logger.Error("could not read metadata", zap.Error(err))
			return nil, err
		}
		return b, nil
	case legacyCurrentlyDispatchedItemsKey, legacyReadIndexKey, legacyWriteIndexKey:
		return nil, nil
	default:
		message := <-c.queue.PeekChan()
		// register callback for consumption
		c.callbacks[key] = message.ConsumeCallback
		return message.Payload(), nil
	}
}

func (c *client) Set(_ context.Context, key string, value []byte) error {
	switch key {
	case metadataKey:
		return c.persistMetaData(value)
	case legacyCurrentlyDispatchedItemsKey, legacyReadIndexKey, legacyWriteIndexKey:
		return nil
	default:
		return c.queue.Put(value)
	}
}

func (c *client) Delete(_ context.Context, key string) error {
	switch key {
	case metadataKey, legacyCurrentlyDispatchedItemsKey, legacyReadIndexKey, legacyWriteIndexKey:
		return nil
	default:
		if callback, ok := c.callbacks[key]; ok {
			callback()
			delete(c.callbacks, key)
		} else {
			c.logger.Error("Cannot find consumption callback", zap.String("key", key))
		}
		return nil
	}
}

func (c *client) Batch(ctx context.Context, ops ...*storage.Operation) error {
	var errs []error
	for _, op := range ops {
		switch op.Type {
		case storage.Set:
			errs = append(errs, c.Set(ctx, op.Key, op.Value))
		case storage.Get:
			var err error
			op.Value, err = c.Get(ctx, op.Key)
			if err != nil {
				errs = append(errs, err)
			}
			continue
		case storage.Delete:
			errs = append(errs, c.Delete(ctx, op.Key))
		}
	}
	return errors.Join(errs...)
}

func (c *client) Close(_ context.Context) error {
	if c.metadataFile != nil {
		_ = c.metadataFile.Close()
		c.metadataFile = nil
	}
	return c.queue.Close()
}

func (d *diskQueueStorageExtension) GetClient(_ context.Context, _ component.Kind, _ component.ID, storageName string) (storage.Client, error) {
	return &client{
		path:                  d.config.Path,
		name:                  storageName,
		queue:                 newQueue(storageName, d.config.Path, d.config.MaxBytesPerFile, d.config.SyncEvery, d.config.SyncTimeout, d.settings.Logger),
		logger:                d.settings.Logger,
		callbacks:             make(map[string]func(), 10),
		metadataTruncateEvery: 1000,
	}, nil
}

func (d *diskQueueStorageExtension) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (d *diskQueueStorageExtension) Shutdown(_ context.Context) error {
	return nil
}

func (c *client) persistMetaData(value []byte) error {
	fileName := filepath.Join(c.path, c.name+"-"+metadataKey)
	if c.metadataFile == nil {
		f, err := os.OpenFile(fileName, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0o600)
		if err != nil {
			return err
		}
		c.metadataFile = f
	}
	_, err := c.metadataFile.Write(value)
	if err != nil {
		_ = c.metadataFile.Close()
		c.metadataFile = nil
		return err
	}
	c.metadataWrites++

	if c.metadataWrites%c.metadataTruncateEvery == 0 {
		_ = c.metadataFile.Close()
		c.metadataFile = nil
		c.metadataWrites = 0
	}

	return nil
}

func (c *client) readMetadata() ([]byte, error) {
	fileName := filepath.Join(c.path, c.name+"-"+metadataKey)
	b, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	lastMetadataUpdate := b[bytes.LastIndex(b, []byte(separator)):]
	return lastMetadataUpdate, nil
}
