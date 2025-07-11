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

package discoveryreceiver

import (
	"fmt"
	"regexp"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

var (
	_ component.Config = (*Config)(nil)

	allowedMatchTypes = []string{"regexp", "strict", "expr"}

	receiverCreatorRegexp = regexp.MustCompile(`receiver_creator/`)
	endpointTargetRegexp  = regexp.MustCompile(`{endpoint=[^}]*}/`)
)

type Config struct {
	// Receivers is a mapping of receivers to discover to their receiver creator configs
	// and evaluated metrics and application statements, which are used to determine component status.
	Receivers map[component.ID]ReceiverEntry `mapstructure:"receivers"`
	// The configured Observer extensions from which to receive Endpoint events.
	// Must implement the observer.Observable interface.
	WatchObservers []component.ID `mapstructure:"watch_observers"`
	// Whether to include the receiver config as a base64-encoded "discovery.receiver.config"
	// resource attribute string value. Will also contain the configured observer that
	// produced the endpoint leading to receiver creation in `watch_observers`.
	// Warning: these values will include the literal receiver subconfig from the parent Collector config.
	// The feature provides no secret redaction and its output is easily decodable into plaintext.
	EmbedReceiverConfig bool `mapstructure:"embed_receiver_config"`
	// The duration to maintain "removed" endpoints since their last updated timestamp.
	CorrelationTTL time.Duration `mapstructure:"correlation_ttl"`
}

// ReceiverEntry is a definition for a receiver instance to instantiate for each Endpoint matching
// the defined rule. Its Config, ResourceAttributes, and Rule will be marshaled to the internal
// Receiver Creator config.
type ReceiverEntry struct {
	Config             map[string]any    `mapstructure:"config"`
	Status             *Status           `mapstructure:"status"`
	ResourceAttributes map[string]string `mapstructure:"resource_attributes"`
	Rule               Rule              `mapstructure:"rule"`
	ServiceType        string            `mapstructure:"service_type"`
}

// Status defines the Match rules for applicable app and telemetry sources.
// The first matching rule determines status of the endpoint.
// At this time only Metrics and zap logger Statements status source types are supported.
type Status struct {
	Metrics    []Match `mapstructure:"metrics"`
	Statements []Match `mapstructure:"statements"`
}

// Match defines the rules for the desired match type and resulting log record
// content emitted by the Discovery receiver
type Match struct {
	Status  discovery.StatusType `mapstructure:"status"`
	Message string               `mapstructure:"message"`
	Strict  string               `mapstructure:"strict"`
	Regexp  string               `mapstructure:"regexp"`
	Expr    string               `mapstructure:"expr"`
}

func (cfg *Config) Validate() error {
	var err error
	for rName, rEntry := range cfg.Receivers {
		name := rName.String()
		if rName.Type() == component.MustNewType("receiver_creator") {
			err = multierr.Combine(err, fmt.Errorf("receiver %q validation failure: receiver cannot be a receiver_creator", name))
			continue
		}
		// These check reserved separators used by the Receiver Creator by which we
		// obtain receiver component.ID and observer.EndpointID. If they are used
		// directly in config we won't be able to determine the ids.
		for _, re := range []*regexp.Regexp{receiverCreatorRegexp, endpointTargetRegexp} {
			if re.MatchString(name) {
				err = multierr.Combine(err, fmt.Errorf("receiver %q validation failure: receiver name cannot contain %q", name, re.String()))
			}
		}
		if e := rEntry.validate(); e != nil {
			err = multierr.Combine(err, fmt.Errorf("receiver %q validation failure: %w", name, e))
		}
	}

	if len(cfg.WatchObservers) == 0 {
		err = multierr.Combine(err, fmt.Errorf("`watch_observers` must be defined and include at least one configured observer extension"))
	}

	return err
}

func (re *ReceiverEntry) validate() error {
	if re.ServiceType == "" {
		return fmt.Errorf("`service_type` must be defined for each receiver")
	}
	return re.Status.validate()
}

func (s *Status) validate() error {
	if s == nil {
		return fmt.Errorf("`status` must be defined and contain at least one `metrics` or `statements` mapping")
	}

	if len(s.Metrics) == 0 && len(s.Statements) == 0 {
		return fmt.Errorf("`status` must contain at least one `metrics` or `statements` list")
	}

	var err error
	statusSources := []struct {
		sourceType string
		matches    []Match
	}{{"metrics", s.Metrics}, {"statements", s.Statements}}
	for _, statusSource := range statusSources {
		for _, statement := range statusSource.matches {
			if ok, e := discovery.IsValidStatus(statement.Status); !ok {
				err = multierr.Combine(err, fmt.Errorf(`"%s" status match validation failed: %w`, statusSource.sourceType, e))
				continue
			}
			var matchTypes []string
			if statement.Strict != "" {
				matchTypes = append(matchTypes, "strict")
			}
			if statement.Regexp != "" {
				matchTypes = append(matchTypes, "regexp")
			}
			if statement.Expr != "" {
				matchTypes = append(matchTypes, "expr")
			}
			if len(matchTypes) != 1 {
				err = multierr.Combine(err, fmt.Errorf(
					`"%s" status match validation failed. Must provide one of %v but received %v`, statusSource.sourceType, allowedMatchTypes, matchTypes,
				))
			}
		}
	}
	return err
}

// receiverCreatorFactoryAndConfig will embed the applicable receiver creator fields in a new receiver creator config
// suitable for being used to create a receiver instance by the returned factory.
func (cfg *Config) receiverCreatorFactoryAndConfig() (receiver.Factory, component.Config, error) {
	receiverCreatorFactory := receivercreator.NewFactory()
	receiverCreatorDefaultConfig := receiverCreatorFactory.CreateDefaultConfig()
	receiverCreatorConfig, ok := receiverCreatorDefaultConfig.(*receivercreator.Config)
	if !ok {
		return nil, nil, fmt.Errorf("failed to coerce to receivercreator.Config")
	}

	receiverCreatorConfig.WatchObservers = cfg.WatchObservers

	receiversConfig := cfg.receiverCreatorReceiversConfig()
	receiverTemplates := confmap.NewFromStringMap(map[string]any{"receivers": receiversConfig})
	if err := receiverCreatorConfig.Unmarshal(receiverTemplates); err != nil {
		return nil, nil, fmt.Errorf("failed unmarshaling discoveryreceiver receiverTemplates into receiver_creator config: %w", err)
	}

	return receiverCreatorFactory, receiverCreatorConfig, nil
}

// receiverCreatorReceiversConfig produces the actual config string map used by the receiver creator config unmarshaler.
func (cfg *Config) receiverCreatorReceiversConfig() map[string]any {
	receiversConfig := map[string]any{}
	for receiverID, rEntry := range cfg.Receivers {
		resourceAttributes := map[string]string{}
		for k, v := range rEntry.ResourceAttributes {
			resourceAttributes[k] = v
		}
		resourceAttributes[discovery.ReceiverNameAttr] = receiverID.Name()
		resourceAttributes[discovery.ReceiverTypeAttr] = receiverID.Type().String()
		resourceAttributes[discovery.EndpointIDAttr] = "`id`"

		rEntryMap := map[string]any{}
		rEntryMap["rule"] = rEntry.Rule.String()
		rEntryMap["config"] = rEntry.Config
		rEntryMap["resource_attributes"] = resourceAttributes
		receiversConfig[receiverID.String()] = rEntryMap
	}

	return receiversConfig
}
