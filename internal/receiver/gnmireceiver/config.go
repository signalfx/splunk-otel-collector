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
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configtls"
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

// Target types accepted from devices registering over the gRPC tunnel.
//   - GNMI_GNOI is production.
//   - GNMI_GNOI_INSECURE is Cisco's testing-only variant
//     (dials the device's insecure gNMI port inside the still-TLS tunnel).
const (
	targetTypeGNMIGNOI         = "GNMI_GNOI"
	targetTypeGNMIGNOIInsecure = "GNMI_GNOI_INSECURE"
)

// Config defines the configuration for the gNMI receiver.
type Config struct {
	// Targets is the list of gNMI devices to subscribe to.
	Targets []TargetConfig `mapstructure:"targets"`

	// GNMIReverse configures an optional gNMIReverse gRPC server for devices that
	// stream gNMI Subscribe or Get responses to the receiver. An empty endpoint disables it.
	GNMIReverse *GNMIReverseConfig `mapstructure:"gnmireverse"`

	// Tunnel configures an optional OpenConfig gRPC tunnel server for devices
	// that dial out to the receiver and register a tunnel target. The receiver
	// then dials back through the tunnel and runs a normal gNMI Subscribe.
	// An empty endpoint disables it.
	Tunnel *TunnelConfig `mapstructure:"tunnel"`

	// YangModules is a list of YANG module files and/or directories used to
	// derive each leaf's metric type (gauge/sum), unit, and enum/identityref
	// value set directly from the schema. Directories are searched
	// recursively for "*.yang" files and are also added as import/include
	// search paths, so a module that only exists to satisfy another module's
	// "import" statement does not need to be listed explicitly. When set,
	// "default"/"overrides" and their "type" become optional: an explicit
	// "overrides" entry still wins outright over the schema, and "default"
	// still fills in whatever the schema left blank (e.g. "unit", which most
	// OpenConfig models omit). Optional.
	YangModules []string `mapstructure:"yang_modules"`
}

// TargetConfig defines connectivity, authentication, and subscriptions for a
// single gNMI target.
type TargetConfig struct {
	// Username and Password are sent as gNMI gRPC metadata.
	Username configopaque.String `mapstructure:"username"`
	Password configopaque.String `mapstructure:"password"`
	// Encoding is the gNMI encoding to request: "proto", "json", or "json_ietf".
	// Defaults to "proto" when omitted.
	Encoding string `mapstructure:"encoding"`
	// Subscriptions is the list of paths to subscribe to on this target.
	Subscriptions []SubscriptionConfig    `mapstructure:"subscriptions"`
	ClientConfig  configgrpc.ClientConfig `mapstructure:",squash"`
	// Redial is the delay before reconnecting after a session failure.
	// Set to 0 to disable automatic reconnection.
	Redial time.Duration `mapstructure:"redial"`
}

// GNMIReverseConfig defines the optional receiver-side gNMI server.
type GNMIReverseConfig struct {
	Subscriptions           []SubscriptionConfig `mapstructure:"subscriptions"`
	configgrpc.ServerConfig `mapstructure:",squash"`
}

// TunnelConfig defines the optional receiver-hosted gRPC tunnel server and
// the subscription set applied to every target that registers over it.
type TunnelConfig struct {
	Username                configopaque.String    `mapstructure:"username"`
	Password                configopaque.String    `mapstructure:"password"`
	Encoding                string                 `mapstructure:"encoding"`
	ClientTLS               configtls.ClientConfig `mapstructure:"client_tls"`
	Subscriptions           []SubscriptionConfig   `mapstructure:"subscriptions"`
	TargetTypes             []string               `mapstructure:"target_types"`
	configgrpc.ServerConfig `mapstructure:",squash"`
	Redial                  time.Duration `mapstructure:"redial"`
}

func (t *TunnelConfig) targetConfig(targetType string) TargetConfig {
	tc := NewDefaultTargetConfig()
	tc.Username = t.Username
	tc.Password = t.Password
	tc.Encoding = t.Encoding
	tc.Redial = t.Redial
	tc.Subscriptions = t.Subscriptions
	if targetType == targetTypeGNMIGNOIInsecure {
		tc.ClientConfig.TLS = configtls.NewDefaultClientConfig()
		tc.ClientConfig.TLS.Insecure = true
	} else {
		tc.ClientConfig.TLS = t.ClientTLS
	}
	return tc
}

func NewDefaultTargetConfig() TargetConfig {
	return TargetConfig{
		ClientConfig: configgrpc.NewDefaultClientConfig(),
		Encoding:     encodingProto,
		Redial:       defaultRedial,
	}
}

func NewDefaultGNMIReverseConfig() GNMIReverseConfig {
	return GNMIReverseConfig{ServerConfig: configgrpc.NewDefaultServerConfig()}
}

func NewDefaultTunnelConfig() TunnelConfig {
	return TunnelConfig{
		ServerConfig: configgrpc.NewDefaultServerConfig(),
		ClientTLS:    configtls.NewDefaultClientConfig(),
		TargetTypes:  []string{targetTypeGNMIGNOI},
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

func (sc SubscriptionConfig) copy() SubscriptionConfig {
	out := sc
	if sc.Default != nil {
		d := *sc.Default
		out.Default = &d
	}
	if len(sc.Overrides) > 0 {
		overrides := make(map[string]MetricConfig, len(sc.Overrides))
		for leaf, cfg := range sc.Overrides {
			overrides[leaf] = cfg
		}
		out.Overrides = overrides
	}
	return out
}

// MetricConfig declares how a gNMI leaf value is represented as an OTel metric.
type MetricConfig struct {
	// Type is the OTel metric type: "gauge" or "sum".
	Type string `mapstructure:"type"`

	// Unit is the metric unit, ideally in UCUM (e.g. "By", "1", "By/s").
	// Optional.
	Unit string `mapstructure:"unit"`

	// EnumValues declares the closed set of values a string leaf can take.
	// When set, the leaf is emitted as a "_state" metric with one datapoint
	// per declared value (1 for the active value, 0 for the rest).
	// Applies only to leaves that arrive as strings; it has no effect
	// on numeric or boolean values. Optional.
	EnumValues []string `mapstructure:"enum_values"`

	// normalizedEnumValues caches normalizeEnumValue(EnumValues[i]) at parser
	// construction time. Populated by newMetricParser.
	normalizedEnumValues []string
}

var (
	_ component.Config    = (*Config)(nil)
	_ confmap.Validator   = (*Config)(nil)
	_ confmap.Unmarshaler = (*TargetConfig)(nil)
	_ confmap.Validator   = (*TargetConfig)(nil)
	_ confmap.Validator   = (*MetricConfig)(nil)
	_ confmap.Unmarshaler = (*TunnelConfig)(nil)
	_ confmap.Validator   = (*TunnelConfig)(nil)
)

// Unmarshal applies per-target defaults (embedded gRPC client defaults,
// encoding, redial) before decoding user-supplied values.
func (t *TargetConfig) Unmarshal(conf *confmap.Conf) error {
	*t = NewDefaultTargetConfig()
	return conf.Unmarshal(t)
}

// Unmarshal applies tunnel defaults (target types, encoding, redial, tunnel transport,
// and inner-TLS defaults) before decoding user-supplied values.
func (t *TunnelConfig) Unmarshal(conf *confmap.Conf) error {
	*t = NewDefaultTunnelConfig()
	return conf.Unmarshal(t)
}

func (cfg *Config) Validate() error {
	gnmiReverseEndpoint := ""
	if cfg.GNMIReverse != nil {
		gnmiReverseEndpoint = cfg.GNMIReverse.NetAddr.Endpoint
	}
	tunnelEndpoint := ""
	if cfg.Tunnel != nil {
		tunnelEndpoint = cfg.Tunnel.NetAddr.Endpoint
	}
	if len(cfg.Targets) == 0 && gnmiReverseEndpoint == "" && tunnelEndpoint == "" {
		return errors.New("at least one target, a gnmireverse endpoint, or a tunnel endpoint must be specified")
	}
	if gnmiReverseEndpoint != "" && len(cfg.GNMIReverse.Subscriptions) == 0 {
		return errors.New("gnmireverse.subscriptions must contain at least one subscription when gnmireverse.endpoint is set")
	}
	if gnmiReverseEndpoint != "" {
		for i, sub := range cfg.GNMIReverse.Subscriptions {
			prefix := fmt.Sprintf("gnmireverse.subscriptions[%d]", i)
			if sub.Mode != "" {
				return fmt.Errorf("%s.mode is not supported for gnmireverse subscriptions", prefix)
			}
			if sub.SampleInterval != 0 {
				return fmt.Errorf("%s.sample_interval is not supported for gnmireverse subscriptions", prefix)
			}
			if sub.HeartbeatInterval != 0 {
				return fmt.Errorf("%s.heartbeat_interval is not supported for gnmireverse subscriptions", prefix)
			}
			if sub.SuppressRedundant {
				return fmt.Errorf("%s.suppress_redundant is not supported for gnmireverse subscriptions", prefix)
			}
		}
	}
	if len(cfg.YangModules) == 0 {
		if err := validateMetricConfigRequired(cfg.Targets); err != nil {
			return err
		}
		if gnmiReverseEndpoint != "" {
			if err := validateSubscriptionMetricConfig(cfg.GNMIReverse.Subscriptions, "gnmireverse.subscriptions", false); err != nil {
				return err
			}
		}
		if tunnelEndpoint != "" {
			if err := validateSubscriptionMetricConfig(cfg.Tunnel.Subscriptions, "tunnel.subscriptions", true); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateMetricConfigRequired enforces that every leaf has a source of  "type"
// when there is no YANG schema to supply one. It is skipped entirely when yang_modules is set:
// at runtime, resolve() merges an "overrides" (or "default") entry onto the schema result per field,
// so a "type" left blank there is expected to be filled in by the schema, not a validation gap.
func validateMetricConfigRequired(targets []TargetConfig) error {
	for ti := range targets {
		for si, sub := range targets[ti].Subscriptions {
			prefix := fmt.Sprintf("targets[%d].subscriptions[%d]", ti, si)
			if err := validateOneSubscriptionMetricConfig(sub, prefix, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSubscriptionMetricConfig(subscriptions []SubscriptionConfig, prefix string, validateMode bool) error {
	for i, sub := range subscriptions {
		if err := validateOneSubscriptionMetricConfig(sub, fmt.Sprintf("%s[%d]", prefix, i), validateMode); err != nil {
			return err
		}
	}
	return nil
}

func validateOneSubscriptionMetricConfig(sub SubscriptionConfig, prefix string, validateMode bool) error {
	if sub.Default == nil && len(sub.Overrides) == 0 {
		return fmt.Errorf("%s: at least one of \"default\" or \"overrides\" must be specified when yang_modules is not set", prefix)
	}
	if sub.Default != nil && sub.Default.Type == "" {
		return fmt.Errorf("%s.default: type is required when yang_modules is not set", prefix)
	}
	for leaf, override := range sub.Overrides {
		if override.Type == "" {
			return fmt.Errorf("%s.overrides[%q]: type is required when yang_modules is not set", prefix, leaf)
		}
	}

	if sub.Path == "" {
		return errors.New("path is required")
	}
	if !strings.HasPrefix(sub.Path, "/") {
		return fmt.Errorf("path %q must be absolute (start with %q)", sub.Path, "/")
	}

	if validateMode {
		switch sub.Mode {
		case modeSample:
			if sub.SampleInterval <= 0 {
				return errors.New("sample_interval must be > 0 for \"sample\" mode")
			}
		case modeOnChange, modeTargetDefined:
			if sub.SampleInterval != 0 {
				return fmt.Errorf("sample_interval must not be set for %q mode", sub.Mode)
			}
		case "":
			return errors.New("mode is required")
		default:
			return fmt.Errorf("invalid mode %q (supported: %q, %q, %q)",
				sub.Mode, modeSample, modeOnChange, modeTargetDefined)
		}
	}

	if sub.HeartbeatInterval < 0 {
		return errors.New("heartbeat_interval must be >= 0")
	}
	return nil
}

func (t *TargetConfig) Validate() error {
	if t.ClientConfig.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if err := validateEncoding(t.Encoding); err != nil {
		return err
	}
	if err := validateCredentials(t.Username, t.Password); err != nil {
		return err
	}
	if err := validateRedial(t.Redial); err != nil {
		return err
	}

	if len(t.Subscriptions) == 0 {
		return errors.New("at least one subscription must be specified")
	}
	return nil
}

func validateEncoding(encoding string) error {
	switch encoding {
	case encodingProto, encodingJSON, encodingJSONIETF:
		return nil
	case "":
		return errors.New("encoding is required")
	default:
		return fmt.Errorf("invalid encoding %q (supported: %q, %q, %q)",
			encoding, encodingProto, encodingJSON, encodingJSONIETF)
	}
}

func validateCredentials(username, password configopaque.String) error {
	if username == "" && password != "" {
		return errors.New("password requires username to be set")
	}
	return nil
}

func validateRedial(redial time.Duration) error {
	if redial < 0 {
		return errors.New("redial must not be negative (use 0 to disable reconnection)")
	}
	if redial > 0 && redial < time.Second {
		return errors.New("redial must be at least 1s (or 0 to disable reconnection)")
	}
	return nil
}

func (t *TunnelConfig) Validate() error {
	// if endpoint not provided we assume it's disabled
	if t.NetAddr.Endpoint == "" {
		return nil
	}
	if len(t.Subscriptions) == 0 {
		return errors.New("at least one subscription must be specified")
	}
	if len(t.TargetTypes) == 0 {
		return errors.New("target_types must contain at least one target type")
	}
	seen := make(map[string]struct{}, len(t.TargetTypes))
	for _, tt := range t.TargetTypes {
		switch tt {
		case targetTypeGNMIGNOI, targetTypeGNMIGNOIInsecure:
		default:
			return fmt.Errorf("invalid target_types entry %q (supported: %q, %q)",
				tt, targetTypeGNMIGNOI, targetTypeGNMIGNOIInsecure)
		}
		if _, ok := seen[tt]; ok {
			return fmt.Errorf("duplicate target_types entry %q", tt)
		}
		seen[tt] = struct{}{}
	}
	if err := validateEncoding(t.Encoding); err != nil {
		return err
	}
	if err := validateCredentials(t.Username, t.Password); err != nil {
		return err
	}
	if err := validateRedial(t.Redial); err != nil {
		return err
	}
	return nil
}

func (m *MetricConfig) Validate() error {
	switch m.Type {
	case metricTypeGauge, metricTypeSum, "":
	default:
		return fmt.Errorf("invalid type %q (supported: %q, %q)",
			m.Type, metricTypeGauge, metricTypeSum)
	}

	if len(m.EnumValues) > 0 {
		if m.Type == metricTypeSum {
			return fmt.Errorf("enum_values is not supported with type %q (state metrics must be a gauge)", metricTypeSum)
		}
		seen := make(map[string]string, len(m.EnumValues))
		for _, v := range m.EnumValues {
			if strings.TrimSpace(v) == "" {
				return errors.New("enum_values must not contain empty values")
			}
			normalized := normalizeEnumValue(v)
			if orig, ok := seen[normalized]; ok {
				return fmt.Errorf("enum_values contains duplicate values %q and %q", orig, v)
			}
			seen[normalized] = v
		}
	}
	return nil
}

func copySubscriptions(subs []SubscriptionConfig) []SubscriptionConfig {
	out := make([]SubscriptionConfig, len(subs))
	for i, s := range subs {
		out[i] = s.copy()
	}
	return out
}
