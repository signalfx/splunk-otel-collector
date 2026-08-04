// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gnmireceiver

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/confmap"
)

// Supported gNMI encodings.
const (
	encodingProto    = "proto"
	encodingJSON     = "json"
	encodingJSONIETF = "json_ietf"
)

// Supported subscription modes (STREAM subscriptions).
const (
	modeSample        = "sample"
	modeOnChange      = "on_change"
	modeTargetDefined = "target_defined"
)

// Supported OTel metric types for a leaf, matching OpenTelemetry metric data
// type names. "sum" maps to a monotonic Sum, "gauge" maps to a Gauge.
const (
	metricTypeGauge = "gauge"
	metricTypeSum   = "sum"
)

const defaultRedial = 10 * time.Second

// Config defines the configuration for the gNMI receiver.
type Config struct {
	// Targets is the list of gNMI devices to subscribe to.
	Targets []TargetConfig `mapstructure:"targets"`
}

// TargetConfig defines connectivity, authentication, and subscriptions for a
// single gNMI target.
//
//nolint:govet // fieldalignment: suboptimal layout is inherited from the embedded configgrpc.ClientConfig
type TargetConfig struct {
	ClientConfig configgrpc.ClientConfig `mapstructure:",squash"`

	// Username and Password are sent as gNMI gRPC metadata.
	Username configopaque.String `mapstructure:"username"`
	Password configopaque.String `mapstructure:"password"`

	// Encoding is the gNMI encoding to request: "proto", "json", or "json_ietf".
	// Defaults to "proto" when omitted.
	Encoding string `mapstructure:"encoding"`

	// Subscriptions is the list of paths to subscribe to on this target.
	Subscriptions []SubscriptionConfig `mapstructure:"subscriptions"`

	// Redial is the delay before reconnecting after a session failure.
	// Set to 0 to disable automatic reconnection.
	Redial time.Duration `mapstructure:"redial"`
}

func NewDefaultTargetConfig() TargetConfig {
	return TargetConfig{
		ClientConfig: configgrpc.NewDefaultClientConfig(),
		Encoding:     encodingProto,
		Redial:       defaultRedial,
	}
}

// SubscriptionConfig defines a single gNMI path subscription.
type SubscriptionConfig struct {
	// Default is the metric type/unit applied to every leaf under this
	// subscription unless a more specific entry in Overrides matches. Optional.
	Default *MetricConfig `mapstructure:"default"`

	// Overrides maps a leaf name (the final gNMI path element, e.g. "in-octets")
	// to its metric type/unit, taking precedence over Default. Optional.
	Overrides map[string]MetricConfig `mapstructure:"overrides"`

	// Path is the gNMI path to subscribe to, e.g.
	// "/interfaces/interface/state/counters".
	Path string `mapstructure:"path"`

	// Origin is the YANG model origin (e.g. "openconfig"). Optional.
	Origin string `mapstructure:"origin"`

	// Mode is the subscription mode: "sample", "on_change", or "target_defined".
	Mode string `mapstructure:"mode"`

	// SampleInterval is the sampling period for "sample" mode. Required (> 0)
	// when Mode is "sample".
	SampleInterval time.Duration `mapstructure:"sample_interval"`

	// HeartbeatInterval forces an update at this interval even if the value has
	// not changed. Must be > 0 when set. Optional.
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`

	// SuppressRedundant avoids sending unchanged values. Optional.
	SuppressRedundant bool `mapstructure:"suppress_redundant"`
}

// MetricConfig declares how a gNMI leaf value is represented as an OTel metric.
type MetricConfig struct {
	// Type is the OTel metric type: "gauge" or "sum".
	Type string `mapstructure:"type"`

	// Unit is the metric unit, ideally in UCUM (e.g. "By", "1", "By/s").
	// Optional.
	Unit string `mapstructure:"unit"`
}

var (
	_ component.Config    = (*Config)(nil)
	_ confmap.Validator   = (*Config)(nil)
	_ confmap.Unmarshaler = (*TargetConfig)(nil)
	_ confmap.Validator   = (*TargetConfig)(nil)
	_ confmap.Validator   = (*SubscriptionConfig)(nil)
	_ confmap.Validator   = (*MetricConfig)(nil)
)

// Unmarshal applies per-target defaults (embedded gRPC client defaults,
// encoding, redial) before decoding user-supplied values.
func (t *TargetConfig) Unmarshal(conf *confmap.Conf) error {
	*t = NewDefaultTargetConfig()
	return conf.Unmarshal(t)
}

func (cfg *Config) Validate() error {
	if len(cfg.Targets) == 0 {
		return errors.New("at least one target must be specified")
	}
	return nil
}

func (t *TargetConfig) Validate() error {
	if t.ClientConfig.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	switch t.Encoding {
	case encodingProto, encodingJSON, encodingJSONIETF:
	case "":
		return errors.New("encoding is required")
	default:
		return fmt.Errorf("invalid encoding %q (supported: %q, %q, %q)",
			t.Encoding, encodingProto, encodingJSON, encodingJSONIETF)
	}

	if t.Redial > 0 && t.Redial < time.Second {
		return errors.New("redial must be at least 1s (or 0 to disable reconnection)")
	}

	if len(t.Subscriptions) == 0 {
		return errors.New("at least one subscription must be specified")
	}
	return nil
}

func (s *SubscriptionConfig) Validate() error {
	if s.Path == "" {
		return errors.New("path is required")
	}
	switch s.Mode {
	case modeSample:
		if s.SampleInterval <= 0 {
			return errors.New("sample_interval must be > 0 for \"sample\" mode")
		}
	case modeOnChange, modeTargetDefined:
	case "":
		return errors.New("mode is required")
	default:
		return fmt.Errorf("invalid mode %q (supported: %q, %q, %q)",
			s.Mode, modeSample, modeOnChange, modeTargetDefined)
	}

	if s.HeartbeatInterval < 0 {
		return errors.New("heartbeat_interval must be >= 0")
	}

	if s.Default == nil && len(s.Overrides) == 0 {
		return errors.New("at least one of \"default\" or \"overrides\" must be specified")
	}
	return nil
}

func (m *MetricConfig) Validate() error {
	switch m.Type {
	case metricTypeGauge, metricTypeSum:
	case "":
		return errors.New("type is required")
	default:
		return fmt.Errorf("invalid type %q (supported: %q, %q)",
			m.Type, metricTypeGauge, metricTypeSum)
	}
	return nil
}
