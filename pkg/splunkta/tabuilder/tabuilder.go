// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package tabuilder reads a technical addon's configuration files and builds
// the per-input logs receivers and per-output logs exporters. It is the single
// source of truth for the stanza -> component mapping, shared by the standalone
// collector runner (internal/collector) and the OTel plugins
// (pkg/splunkinputsreceiver).
package tabuilder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/receiver"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/batchreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/monitorreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/scriptreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/tcpreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/udpreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/receiver/wineventlogreceiver"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
)

// ResolveSplunkHome returns baseDir if set, otherwise falls back to $SPLUNK_HOME.
func ResolveSplunkHome(baseDir string) (string, error) {
	if baseDir != "" {
		return baseDir, nil
	}
	if home := os.Getenv("SPLUNK_HOME"); home != "" {
		return home, nil
	}
	return "", fmt.Errorf("base_dir is not set and $SPLUNK_HOME is not defined")
}

// CreateReceivers builds a logs receiver for every enabled input stanza,
// dispatching by the stanza name's input kind. Stanzas with disabled=1 or
// unsupported kinds are skipped silently.
func CreateReceivers(ctx context.Context, inputs []conf.Input, transforms []conf.Transform, props []conf.Prop, baseDir string, next consumer.Logs, telemetrySettings component.TelemetrySettings) ([]receiver.Logs, error) {
	var receivers []receiver.Logs
	for _, input := range inputs {
		if input.Configuration.Stanza.IsDisabled() {
			continue
		}
		l, err := CreateReceiver(ctx, baseDir, next, input, transforms, props, telemetrySettings)
		if err != nil {
			return nil, fmt.Errorf("failed to create receiver %q: %w", input.Configuration.Stanza.Name, err)
		}
		if l == nil {
			continue
		}
		receivers = append(receivers, l)
	}
	return receivers, nil
}

// CreateReceiver builds a single logs receiver for an input stanza.
// Returns nil receiver for unsupported input kinds; the caller is responsible for logging.
func CreateReceiver(ctx context.Context, baseDir string, next consumer.Logs, input conf.Input, transforms []conf.Transform, props []conf.Prop, telemetrySettings component.TelemetrySettings) (receiver.Logs, error) {
	parsed, err := stanza.ParseName(input.Configuration.Stanza.Name)
	if err != nil {
		return nil, err
	}
	switch parsed.Kind {
	case "script", "":
		f := scriptreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), &scriptreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "batch":
		f := batchreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), batchreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "monitor":
		f := monitorreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), monitorreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "wineventlog":
		f := wineventlogreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), wineventlogreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "tcp":
		f := tcpreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), tcpreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	case "udp":
		f := udpreceiver.NewFactory()
		return f.CreateLogs(ctx, settings(f, parsed.Target, telemetrySettings), udpreceiver.Config{
			Input:      input,
			BaseDir:    baseDir,
			Transforms: transforms,
			Props:      props,
		}, next)
	default:
		return nil, nil
	}
}

func settings(f receiver.Factory, path string, telemetrySettings component.TelemetrySettings) receiver.Settings {
	return receiver.Settings{
		ID:                component.MustNewIDWithName(f.Type().String(), path),
		TelemetrySettings: telemetrySettings,
	}
}

func confFilePaths(dirs []string, filename string) []string {
	paths := make([]string, len(dirs))
	for i, dir := range dirs {
		paths[i] = filepath.Join(dir, filename)
	}
	return paths
}

// DiscoverTAs returns splunk_ta_* directories under splunkHome/etc/apps.
func DiscoverTAs(splunkHome string) ([]string, error) {
	appsDir := filepath.Join(splunkHome, "etc", "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("tabuilder: failed to scan %s: %w", appsDir, err)
	}
	var taDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToLower(entry.Name()), "splunk_ta_") {
			taDirs = append(taDirs, filepath.Join(appsDir, entry.Name()))
		}
	}
	return taDirs, nil
}

func splunkHomeDirs(splunkHome string) []string {
	taDirs, _ := DiscoverTAs(splunkHome)
	etcDir := filepath.Join(splunkHome, "etc")

	dirs := []string{filepath.Join(etcDir, "system", "default")}
	for _, ta := range taDirs {
		dirs = append(dirs, filepath.Join(ta, "default"))
	}
	for _, ta := range taDirs {
		dirs = append(dirs, filepath.Join(ta, "local"))
	}
	dirs = append(dirs, filepath.Join(etcDir, "system", "local"))

	return dirs
}

func systemDirs(splunkHome string) []string {
	etcDir := filepath.Join(splunkHome, "etc")
	return []string{
		filepath.Join(etcDir, "system", "default"),
		filepath.Join(etcDir, "system", "local"),
	}
}

func taDirsWithSystem(splunkHome, taDir string) []string {
	etcDir := filepath.Join(splunkHome, "etc")
	return []string{
		filepath.Join(etcDir, "system", "default"),
		filepath.Join(taDir, "default"),
		filepath.Join(taDir, "local"),
		filepath.Join(etcDir, "system", "local"),
	}
}

func readConfFiles(paths []string) ([][]byte, error) {
	var payloads [][]byte
	for _, path := range paths {
		//nolint:gosec // G304: paths are constructed by filepath.Join from validated dirs
		b, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, b)
	}
	return payloads, nil
}

// WatchDirs returns the directories to watch for filesystem changes under splunkHome:
// etc itself (to detect apps/ being created), etc/apps (for TA add/remove),
// and the system conf layers.
func WatchDirs(splunkHome string) []string {
	etcDir := filepath.Join(splunkHome, "etc")
	return []string{
		etcDir, // watch etc itself to detect apps/ being created
		filepath.Join(etcDir, "apps"),
		filepath.Join(etcDir, "system", "default"),
		filepath.Join(etcDir, "system", "local"),
	}
}

// ConfDirs returns the Splunk btool conf search path for splunkHome.
func ConfDirs(splunkHome string) []string {
	return splunkHomeDirs(splunkHome)
}

// SystemDirs returns the system conf directories for splunkHome in precedence order.
func SystemDirs(splunkHome string) []string {
	return systemDirs(splunkHome)
}

// ConfDirsWithSystem returns the Splunk btool conf search path for a single TA merged with system config.
func ConfDirsWithSystem(splunkHome, taDir string) []string {
	return taDirsWithSystem(splunkHome, taDir)
}

// ReadInputs discovers and merges inputs.conf files from the given search
// directories. Use ConfDirs or ConfDirsWithSystem to build the dirs slice.
// Returns nil (no error) when no inputs.conf files are found.
func ReadInputs(dirs []string) ([]conf.Input, error) {
	var layers [][]conf.Input
	for _, path := range confFilePaths(dirs, "inputs.conf") {
		//nolint:gosec // G304: paths are constructed by filepath.Join from validated dirs
		b, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		appDir := filepath.Dir(filepath.Dir(path)) // strip /default or /local
		inputs, err := conf.ReadInput(b, appDir)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		layers = append(layers, inputs)
	}
	return conf.MergeInputs(layers), nil
}

// ReadSystemInputs returns inputs.conf stanzas defined only in etc/system,
// excluding any stanza already owned by a TA. These are run once globally.
func ReadSystemInputs(splunkHome string) ([]conf.Input, error) {
	systemInputs, err := ReadInputs(systemDirs(splunkHome))
	if err != nil {
		return nil, err
	}
	if len(systemInputs) == 0 {
		return nil, nil
	}

	taDirs, err := DiscoverTAs(splunkHome)
	if err != nil {
		return nil, err
	}

	// Collect stanza names owned by any TA so we can exclude them.
	taOwned := map[string]struct{}{}
	for _, taDir := range taDirs {
		inputs, err := ReadInputs([]string{filepath.Join(taDir, "default"), filepath.Join(taDir, "local")})
		if err != nil {
			return nil, err
		}
		for _, input := range inputs {
			taOwned[input.Configuration.Stanza.Name] = struct{}{}
		}
	}

	filtered := systemInputs[:0]
	for _, input := range systemInputs {
		if _, owned := taOwned[input.Configuration.Stanza.Name]; !owned {
			filtered = append(filtered, input)
		}
	}
	return filtered, nil
}

// ReadInputsForTA merges inputs.conf for a single TA with the system conf
// layers, returning only stanzas defined in the TA's own dirs. System layers
// contribute parameter overrides but cannot inject new stanzas.
func ReadInputsForTA(splunkHome, taDir string) ([]conf.Input, error) {
	taDirs := []string{filepath.Join(taDir, "default"), filepath.Join(taDir, "local")}
	allDirs := taDirsWithSystem(splunkHome, taDir)

	// First pass: collect stanza names defined in the TA's own dirs.
	taInputs, err := ReadInputs(taDirs)
	if err != nil {
		return nil, err
	}
	if len(taInputs) == 0 {
		return nil, nil
	}
	taDefined := make(map[string]struct{}, len(taInputs))
	for _, input := range taInputs {
		taDefined[input.Configuration.Stanza.Name] = struct{}{}
	}

	// Second pass: merge all layers (including system) for full precedence.
	merged, err := ReadInputs(allDirs)
	if err != nil {
		return nil, err
	}

	// Filter to only stanzas the TA itself defines.
	filtered := merged[:0]
	for _, input := range merged {
		if _, ok := taDefined[input.Configuration.Stanza.Name]; ok {
			filtered = append(filtered, input)
		}
	}
	return filtered, nil
}

// ReadTransforms discovers and merges transforms.conf files from the given
// search directories. Returns nil (no error) when absent.
func ReadTransforms(dirs []string) ([]conf.Transform, error) {
	payloads, err := readConfFiles(confFilePaths(dirs, "transforms.conf"))
	if err != nil {
		return nil, err
	}
	var layers [][]conf.Transform
	for _, b := range payloads {
		transforms, err := conf.ReadTransforms(b)
		if err != nil {
			return nil, err
		}
		layers = append(layers, transforms)
	}
	return conf.MergeTransforms(layers), nil
}

// ReadProps discovers and merges props.conf files from the given search
// directories. Returns nil (no error) when absent.
func ReadProps(dirs []string) ([]conf.Prop, error) {
	payloads, err := readConfFiles(confFilePaths(dirs, "props.conf"))
	if err != nil {
		return nil, err
	}
	var layers [][]conf.Prop
	for _, b := range payloads {
		props, err := conf.ReadProps(b)
		if err != nil {
			return nil, err
		}
		layers = append(layers, props)
	}
	return conf.MergeProps(layers), nil
}

// ReadOutputs merges outputs.conf across $SPLUNK_HOME using standard Splunk
// precedence. Use HTTPOut (or future TCPOut, etc.) to extract a specific type.
func ReadOutputs(splunkHome string) (conf.ConfMap, error) {
	payloads, err := readConfFiles(confFilePaths(ConfDirs(splunkHome), "outputs.conf"))
	if err != nil {
		return nil, err
	}
	return conf.ParseAndMergeConf(payloads)
}

// ReadOutputGroups merges outputs.conf across $SPLUNK_HOME using standard
// Splunk precedence and returns every output stanza.
func ReadOutputGroups(splunkHome string) ([]conf.Output, error) {
	merged, err := ReadOutputs(splunkHome)
	if err != nil {
		return nil, err
	}
	return conf.OutputGroups(merged)
}

func HTTPOut(merged conf.ConfMap) (*conf.Output, error) {
	return conf.HTTPOut(merged)
}

// CreateExporter builds a logs exporter from a merged outputs conf map.
func CreateExporter(merged conf.ConfMap, logger *zap.Logger, telemetrySettings component.TelemetrySettings) (exporter.Logs, error) {
	output, err := conf.HTTPOut(merged)
	if err != nil {
		return nil, err
	}
	return CreateOutputExporter(output, logger, telemetrySettings)
}

// CreateOutputExporter builds a logs exporter from one built-in output stanza.
// Returns nil exporter for unsupported output kinds; the caller is responsible for logging.
func CreateOutputExporter(output *conf.Output, logger *zap.Logger, telemetrySettings component.TelemetrySettings) (exporter.Logs, error) {
	parsed, err := stanza.ParseOutputName(output.Configuration.Stanza.Name)
	if err != nil {
		return nil, err
	}
	if parsed.Kind != "httpout" {
		return nil, nil
	}
	return newHECExporter(output, logger, telemetrySettings)
}

func newHECExporter(o *conf.Output, logger *zap.Logger, telemetrySettings component.TelemetrySettings) (exporter.Logs, error) {
	f := splunkhecexporter.NewFactory()
	cfg := f.CreateDefaultConfig().(*splunkhecexporter.Config)
	cfg.ClientConfig.Endpoint = outputParam(o, "uri")
	cfg.Token = configopaque.String(outputParam(o, "httpEventCollectorToken"))
	cfg.ClientConfig.TLS = configtls.ClientConfig{InsecureSkipVerify: true} // TODO: wire sslVerifyServerCert from outputs.conf
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	s := exporter.Settings{
		ID: component.MustNewID(f.Type().String()),
		TelemetrySettings: component.TelemetrySettings{
			Logger:         logger,
			TracerProvider: tracenoop.NewTracerProvider(),
			MeterProvider:  metricnoop.NewMeterProvider(),
		},
	}
	if telemetrySettings.Logger != nil {
		s.TelemetrySettings = telemetrySettings
	}
	return f.CreateLogs(context.Background(), s, cfg)
}

func outputParam(o *conf.Output, name string) string {
	if param := o.Configuration.Stanza.Params.Get(name); param != nil {
		return param.Value
	}
	return ""
}
