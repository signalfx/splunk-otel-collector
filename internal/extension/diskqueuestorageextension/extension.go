package diskqueuestorageextension

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/signalfx/splunk-otel-collector/internal/extension/diskqueuestorageextension/internal"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/xextension/storage"
)

const metadataKey = "qmv0"

var _ storage.Extension = (*diskQueueStorageExtension)(nil)

func newDiskQueueStorageExtension(settings extension.Settings) extension.Extension {
	return &diskQueueStorageExtension{}
}

type diskQueueStorageExtension struct {
	config   *Config
	settings extension.Settings
}

type client struct {
	path  string
	queue internal.Interface
}

func (c *client) Get(_ context.Context, key string) ([]byte, error) {
	if key == metadataKey {
		return c.readMetadata()
	} else {
		b := <-c.queue.PeekChan()
		return b, nil
	}
}

func (c *client) Set(_ context.Context, key string, value []byte) error {
	if key == metadataKey {
		return c.persistMetaData(value)
	} else {
		return c.queue.Put(value)
	}
}

func (c *client) Delete(_ context.Context, key string) error {
	if key == metadataKey {
		return nil
	}
	<-c.queue.ReadChan()
	return nil
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
		queue: internal.New(storageName, d.config.Path, d.config.MaxBytesPerFile, d.config.SyncEvery, d.config.SyncTimeout, d.settings.Logger),
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
