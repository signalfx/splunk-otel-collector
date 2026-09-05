// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/splunk/tarunner/pkg/splunkta/conf"
	"github.com/splunk/tarunner/pkg/splunkta/stanza"
	"github.com/splunk/tarunner/pkg/splunkta/tabuilder"
)

type (
	Input         = conf.Input
	Configuration = conf.Configuration
	Stanza        = conf.Stanza
	Params        = conf.Params
	Param         = conf.Param
	Prop          = conf.Prop
	Transform     = conf.Transform
	FieldAlias    = conf.FieldAlias
	PropType      = conf.PropType
)

// ReceiverRequest is passed to a sub-receiver factory for one inputs.conf
// stanza. Path is the parsed target from the stanza name. Empty-kind stanzas
// are dispatched to the "script" sub-receiver.
type ReceiverRequest struct {
	BaseDir    string
	Path       string
	Input      Input
	Transforms []Transform
	Props      []Prop
}

// SubReceiverFactory creates a logs receiver for one inputs.conf stanza kind.
//
// Scheme returns the stanza kind to match. Kinds are matched case-sensitively,
// matching Splunk UF behavior. Returning "script" handles both script:// stanzas and
// empty-kind modular input stanzas.
type SubReceiverFactory interface {
	Scheme() string
	CreateLogs(context.Context, receiver.Settings, ReceiverRequest, consumer.Logs) (receiver.Logs, error)
}

// Option configures the splunk_inputs factory.
type Option func(*factoryOptions)

// WithSubReceiver registers a sub-receiver factory by Scheme. If another
// factory is already registered for the same scheme, it is replaced.
func WithSubReceiver(f SubReceiverFactory) Option {
	return func(o *factoryOptions) {
		if f == nil {
			return
		}
		o.subReceivers[f.Scheme()] = f
	}
}

type factoryOptions struct {
	subReceivers map[string]SubReceiverFactory
}

func newFactoryOptions(opts ...Option) factoryOptions {
	options := factoryOptions{
		subReceivers: map[string]SubReceiverFactory{},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func (o factoryOptions) createLogsFunc(_ context.Context, settings receiver.Settings, config component.Config, logs consumer.Logs) (receiver.Logs, error) {
	cfg := config.(Config)

	splunkHome, err := tabuilder.ResolveSplunkHome(cfg.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("splunk_inputs: %w", err)
	}

	return newSplunkInputsReceiver(splunkHome, o, settings, logs), nil
}

// systemKey is a sentinel used in handler.active to track system-only stanzas
// (those defined in etc/system but not owned by any TA).
const systemKey = "\x00system"

func (o factoryOptions) startReceivers(ctx context.Context, host component.Host, splunkHome, taDir string, next consumer.Logs, settings receiver.Settings) ([]receiver.Logs, error) {
	var inputs []Input
	var dirs []string
	var err error
	if taDir == systemKey {
		inputs, err = tabuilder.ReadSystemInputs(splunkHome)
		dirs = tabuilder.SystemDirs(splunkHome)
	} else {
		inputs, err = tabuilder.ReadInputsForTA(splunkHome, taDir)
		dirs = tabuilder.ConfDirsWithSystem(splunkHome, taDir)
	}
	if err != nil {
		return nil, err
	}
	transforms, err := tabuilder.ReadTransforms(dirs)
	if err != nil {
		return nil, err
	}
	props, err := tabuilder.ReadProps(dirs)
	if err != nil {
		return nil, err
	}
	rcvrs, err := o.createReceivers(ctx, inputs, transforms, props, taDir, next, settings)
	if err != nil {
		return nil, err
	}
	var started []receiver.Logs
	for _, r := range rcvrs {
		if err := r.Start(ctx, host); err != nil {
			settings.Logger.Error("splunk_inputs: failed to start receiver",
				zap.String("ta", taDir), zap.Error(err))
			continue
		}
		started = append(started, r)
	}
	return started, nil
}

func (o factoryOptions) createReceivers(ctx context.Context, inputs []Input, transforms []Transform, props []Prop, baseDir string, next consumer.Logs, settings receiver.Settings) ([]receiver.Logs, error) {
	var receivers []receiver.Logs
	for i := range inputs {
		input := inputs[i]
		name := input.Configuration.Stanza.Name
		if input.Configuration.Stanza.IsDisabled() {
			settings.Logger.Info("splunk_inputs: skipping disabled stanza", zap.String("stanza", name))
			continue
		}
		l, err := o.createReceiver(ctx, baseDir, next, input, transforms, props, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create receiver %q: %w", name, err)
		}
		if l == nil {
			settings.Logger.Info("splunk_inputs: skipping unsupported input stanza", zap.String("stanza", name))
			continue
		}
		receivers = append(receivers, l)
	}
	return receivers, nil
}

func (o factoryOptions) createReceiver(ctx context.Context, baseDir string, next consumer.Logs, input Input, transforms []Transform, props []Prop, settings receiver.Settings) (receiver.Logs, error) {
	parsed, err := stanza.ParseName(input.Configuration.Stanza.Name)
	if err != nil {
		return nil, err
	}
	scheme := parsed.Kind
	if scheme == "" {
		scheme = "script"
	}
	if f, ok := o.subReceivers[scheme]; ok {
		return f.CreateLogs(ctx, settings, ReceiverRequest{
			BaseDir:    baseDir,
			Path:       parsed.Target,
			Input:      input,
			Transforms: transforms,
			Props:      props,
		}, next)
	}
	l, err := tabuilder.CreateReceiver(ctx, baseDir, next, input, transforms, props, settings.TelemetrySettings)
	if l == nil && err == nil {
		return nil, fmt.Errorf("unsupported scheme %q", scheme)
	}
	return l, err
}
