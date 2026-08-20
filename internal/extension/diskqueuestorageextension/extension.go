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
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/xextension/storage"

	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension/internal"
)

// special keys used by the persistent queue to store metadata.
const (
	metadataKey = "qmv0"

	// all legacy keys - ignored.
	legacyReadIndexKey                = "ri"
	legacyWriteIndexKey               = "wi"
	legacyCurrentlyDispatchedItemsKey = "di"
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
	queue internal.Interface
	path  string
}

func (c *client) Get(_ context.Context, key string) ([]byte, error) {
	switch key {
	case metadataKey:
		b, err := c.readMetadata()
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			panic(err)
		}
		return b, nil
	case legacyCurrentlyDispatchedItemsKey, legacyReadIndexKey, legacyWriteIndexKey:
		return nil, nil
	default:
		b := <-c.queue.PeekChan()
		return b, nil
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
		<-c.queue.ReadChan()
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
	return c.queue.Close()
}

func (d *diskQueueStorageExtension) GetClient(_ context.Context, _ component.Kind, _ component.ID, storageName string) (storage.Client, error) {
	return &client{
		path:  d.config.Path,
		queue: internal.New(storageName, d.config.Path, d.config.MaxBytesPerFile, d.config.SyncEvery, d.config.SyncTimeout, d.config.CompactInterval, d.settings.Logger),
	}, nil
}

func (d *diskQueueStorageExtension) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (d *diskQueueStorageExtension) Shutdown(_ context.Context) error {
	return nil
}

func (c *client) persistMetaData(value []byte) error {
	var f *os.File
	var err error

	fileName := filepath.Join(c.path, metadataKey)
	f, err = os.CreateTemp("", path.Base(fileName)+"-*")
	if err != nil {
		return err
	}

	_, err = f.Write(value)
	if err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Sync()
	_ = f.Close()

	// atomically rename
	return os.Rename(f.Name(), fileName)
}

func (c *client) readMetadata() ([]byte, error) {
	fileName := filepath.Join(c.path, metadataKey)
	return os.ReadFile(fileName)
}
