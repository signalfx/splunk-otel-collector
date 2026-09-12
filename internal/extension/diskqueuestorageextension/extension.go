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

	callbacksSize = 10000
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
	name                  string
	path                  string
	callbacks             []map[string]func(metadata []byte)
	metadataWrites        int
	metadataTruncateEvery int
}

func (c *client) Get(_ context.Context, key string) ([]byte, error) {
	// ignore old metadata keys
	if key == legacyCurrentlyDispatchedItemsKey || key == legacyWriteIndexKey || key == legacyReadIndexKey {
		return nil, nil
	}
	// this Get function can only retrieve metadata. Everything else is done as a batch.
	if key != metadataKey {
		panic("The disk_queue_storage extension can only be used with the persistent queue.")
	}
	b := c.queue.metadata.metadata.Load().([]byte)
	return b, nil
}

func (c *client) Set(_ context.Context, key string, value []byte) error {
	// this function cannot be called directly. Deletes must be called via Batch.
	if key != metadataKey {
		panic("The disk_queue_storage extension can only be used with the persistent queue.")
	}
	c.queue.metadata.metadata.Store(value)
	return nil
}

func (c *client) Delete(_ context.Context, _ string) error {
	// this function cannot be called directly. Deletes must be called via Batch.
	panic("The disk_queue_storage extension can only be used with the persistent queue.")
}

func (c *client) Batch(_ context.Context, ops ...*storage.Operation) error {
	// we expect that batch operations are a combination of a queue change + writing the metadata.
	// we combine both to persist it as an atomic operation.
	// ignore batch operations regarding
	if len(ops) == 2 && ops[0].Key == legacyReadIndexKey && ops[1].Key == legacyWriteIndexKey {
		return nil
	}
	if len(ops) == 3 && ops[0].Key == legacyReadIndexKey && ops[1].Key == legacyWriteIndexKey && ops[2].Key == legacyCurrentlyDispatchedItemsKey {
		return nil
	}

	var setMetadata *storage.Operation
	var changeOp *storage.Operation
	for _, op := range ops {
		if op.Type == storage.Set && op.Key == metadataKey {
			setMetadata = op
		} else {
			changeOp = op
		}
	}
	if setMetadata == nil || changeOp == nil {
		return errors.New("invalid batch")
	}

	switch changeOp.Type {
	case storage.Set:
		return c.queue.put(setMetadata.Value, changeOp.Value)
	case storage.Get:
		message := <-c.queue.peek(setMetadata.Value)
		// register callback for consumption

		var localCallbackMap map[string]func([]byte)
		for _, callbackMap := range c.callbacks {
			if len(callbackMap) < callbacksSize {
				localCallbackMap = callbackMap
				break
			}
		}
		if localCallbackMap == nil {
			localCallbackMap = make(map[string]func([]byte), callbacksSize)
			c.callbacks = append(c.callbacks, localCallbackMap)
		}
		localCallbackMap[changeOp.Key] = message.consumeCallback
		changeOp.Value = message.payload
		return nil
	case storage.Delete:
		callbackLen := len(c.callbacks)
		for i := callbackLen - 1; i >= 0; i-- {
			cbMap := c.callbacks[i]
			if callback, ok := cbMap[changeOp.Key]; ok {
				callback(setMetadata.Value)
				delete(cbMap, changeOp.Key)
				if len(cbMap) == 0 && i != callbackLen-1 {
					c.callbacks[i] = make(map[string]func([]byte), callbacksSize)
				}
				return nil
			}
		}
		return errors.New("cannot delete " + changeOp.Key)
	}
	return errors.New("invalid operation")
}

func (c *client) Close(_ context.Context) error {
	return c.queue.close()
}

func (d *diskQueueStorageExtension) GetClient(_ context.Context, _ component.Kind, _ component.ID, storageName string) (storage.Client, error) {
	q, err := newQueue(storageName, d.config.Path, d.config.MaxBytesPerFile, d.config.SyncEvery, d.config.SyncTimeout, d.settings.Logger)
	if err != nil {
		return nil, err
	}
	return &client{
		path:   d.config.Path,
		name:   storageName,
		queue:  q,
		logger: d.settings.Logger,
		callbacks: []map[string]func([]byte){
			make(map[string]func([]byte), callbacksSize),
		},
		metadataTruncateEvery: 1000,
	}, nil
}

func (d *diskQueueStorageExtension) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (d *diskQueueStorageExtension) Shutdown(_ context.Context) error {
	return nil
}
