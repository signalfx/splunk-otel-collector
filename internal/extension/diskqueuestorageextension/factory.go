package diskqueuestorageextension

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

const (
	TypeStr = "disk_queue_storage"
)

func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType(TypeStr),
		createDefaultConfig,
		createExtension,
		component.StabilityLevelAlpha,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		MaxBytesPerFile: 10 * 1024 * 1024,
		SyncEvery:       1,
		SyncTimeout:     100 * time.Millisecond,
		CompactInterval: 30 * time.Minute,
	}
}

func createExtension(
	_ context.Context,
	settings extension.Settings,
	_ component.Config,
) (extension.Extension, error) {
	return newDiskQueueStorageExtension(settings), nil
}
