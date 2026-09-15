// Copyright Splunk, Inc.
// SPDX-License-Identifier: Apache-2.0

package splunkinputsreceiver

import (
	"context"
	"fmt"
	"reflect"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/conf"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/stanza"
	"github.com/signalfx/splunk-otel-collector/pkg/splunkta/tabuilder"
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

// receiverSpec is the effective configuration for one TA input stanza. The
// transforms and props are included because they are compiled into the
// receiver's operator pipeline and therefore changing either changes the
// receiver even when the input stanza itself is unchanged.
type receiverSpec struct {
	name       string
	input      Input
	transforms []Transform
	props      []Prop
}

func (s receiverSpec) equal(other receiverSpec) bool {
	return s.name == other.name &&
		reflect.DeepEqual(s.input, other.input) &&
		reflect.DeepEqual(s.transforms, other.transforms) &&
		reflect.DeepEqual(s.props, other.props)
}

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
	specs, err := o.receiverSpecs(splunkHome, taDir)
	if err != nil {
		return nil, err
	}
	started, err := o.startReceiverSpecs(ctx, host, taDir, next, settings, specs)
	if err != nil {
		return nil, err
	}
	rcvrs := make([]receiver.Logs, 0, len(started))
	for _, r := range started {
		rcvrs = append(rcvrs, r.receiver)
	}
	return rcvrs, nil
}

func (o factoryOptions) receiverSpecs(splunkHome, taDir string) ([]receiverSpec, error) {
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

	specs := make([]receiverSpec, 0, len(inputs))
	for _, input := range inputs {
		if input.Configuration.Stanza.IsDisabled() {
			continue
		}
		name := input.Configuration.Stanza.Name
		_, parseErr := stanza.ParseName(name)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse receiver %q: %w", name, parseErr)
		}
		specs = append(specs, receiverSpec{
			name:       name,
			input:      input,
			transforms: transforms,
			props:      props,
		})
	}
	return specs, nil
}

type startedReceiver struct {
	spec     receiverSpec
	receiver receiver.Logs
}

func (o factoryOptions) startReceiverSpecs(ctx context.Context, host component.Host, baseDir string, next consumer.Logs, settings receiver.Settings, specs []receiverSpec) ([]startedReceiver, error) {
	started := make([]startedReceiver, 0, len(specs))
	for _, spec := range specs {
		r, err := o.createReceiver(ctx, baseDir, next, spec.input, spec.transforms, spec.props, settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create receiver %q: %w", spec.name, err)
		}
		if r == nil {
			continue
		}
		if err := r.Start(ctx, host); err != nil {
			settings.Logger.Error("splunk_inputs: failed to start receiver",
				zap.String("ta", baseDir), zap.String("stanza", spec.name), zap.Error(err))
			continue
		}
		started = append(started, startedReceiver{spec: spec, receiver: r})
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
