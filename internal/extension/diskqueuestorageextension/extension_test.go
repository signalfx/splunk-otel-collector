package diskqueuestorageextension

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

var _ component.Host = (*hostWithExtensions)(nil)

type hostWithExtensions struct {
	extensions map[component.ID]component.Component
}

func (h hostWithExtensions) GetExtensions() map[component.ID]component.Component {
	return h.extensions
}

func TestExtensionAsPersistentQueue(t *testing.T) {
	f := otlpexporter.NewFactory()
	cfg := f.CreateDefaultConfig().(*otlpexporter.Config)
	extId := component.MustNewIDWithName("disk_queue_storage", "my")
	cfg.QueueConfig.GetOrInsertDefault().StorageID = &extId
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	exporterSettings := exportertest.NewNopSettings(component.MustNewType("otlp"))
	exporterSettings.Logger = logger
	l, err := f.CreateLogs(t.Context(), exporterSettings, cfg)
	require.NoError(t, err)
	extConfig := createDefaultConfig().(*Config)
	extConfig.Path = t.TempDir()
	extensionSettings := extensiontest.NewNopSettings(component.MustNewType("disk_queue_storage"))
	extensionSettings.Logger = logger
	require.NoError(t, l.Start(t.Context(), hostWithExtensions{
		extensions: map[component.ID]component.Component{
			extId: &diskQueueStorageExtension{
				config:   extConfig,
				settings: extensionSettings,
			},
		},
	}))
	logs := plog.NewLogs()
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.ConsumeLogs(t.Context(), logs))
	require.NoError(t, l.Shutdown(t.Context()))

}
