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
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

var (
	_ config.Receiver = (*Config)(nil)

	allowedMatchTypes = []string{"regexp", "strict", "expr"}

	receiverCreatorRegexp = regexp.MustCompile(`receiver_creator/`)
	endpointTargetRegexp  = regexp.MustCompile(`{endpoint=[^}]*}/`)
)

type Config struct {
	// Receivers is a mapping of receivers to discover to their receiver creator configs
	// and evaluated metrics and application statements, which are used to determine component status.
	Receivers               map[config.ComponentID]ReceiverEntry `mapstructure:"receivers"`
	config.ReceiverSettings `mapstructure:",squash"`
	// The configured Observer extensions from which to receive Endpoint events.
	// Must implement the observer.Observable interface.
	WatchObservers []config.ComponentID `mapstructure:"watch_observers"`
	// Whether to emit log records for all endpoint activity, consisting of Endpoint
	// content as record attributes.
	LogEndpoints bool `mapstructure:"log_endpoints"`
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
	Rule               string            `mapstructure:"rule"`
}

// Status defines the Match rules for applicable app and telemetry sources.
// At this time only Metrics and zap logger Statements status source types are supported.
type Status struct {
	Metrics    map[discovery.StatusType][]Match `mapstructure:"metrics"`
	Statements map[discovery.StatusType][]Match `mapstructure:"statements"`
}

// Match defines the rules for the desired match type and resulting log record
// content emitted by the Discovery receiver
type Match struct {
	Record    *LogRecord `mapstructure:"log_record"`
	Strict    string     `mapstructure:"strict"`
	Regexp    string     `mapstructure:"regexp"`
	Expr      string     `mapstructure:"expr"`
	FirstOnly bool       `mapstructure:"first_only"`
}

// LogRecord is a definition of the desired plog.LogRecord content to emit for a match.
type LogRecord struct {
	Attributes   map[string]string `mapstructure:"attributes"`
	SeverityText string            `mapstructure:"severity_text"`
	Body         string            `mapstructure:"body"`
}

func (cfg *Config) Validate() error {
	var err error
	for rName, rEntry := range cfg.Receivers {
		name := rName.String()
		if rName.Type() == "receiver_creator" {
			err = multierr.Combine(err, fmt.Errorf("receiver %q validation failure: receiver cannot be a receiver_creator", name))
			continue
		}
		// These check reserved separators used by the Receiver Creator by which we
		// obtain receiver config.ComponentID and observer.EndpointID. If they are used
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

	if cfg.WatchObservers == nil || len(cfg.WatchObservers) == 0 {
		err = multierr.Combine(err, fmt.Errorf("`watch_observers` must be defined and include at least one configured observer extension"))
	}

	return err
}

func (re *ReceiverEntry) validate() error {
	if err := re.Status.validate(); err != nil {
		return err
	}
	return nil
}

func (s *Status) validate() error {
	if s == nil {
		return fmt.Errorf("`status` must be defined and contain at least one `metrics` or `statements` mapping")
	}

	if len(s.Metrics) == 0 && len(s.Statements) == 0 {
		return fmt.Errorf("`status` must contain at least one `metrics` or `statements` mapping with at least one of %v", discovery.StatusTypes)
	}

	var err error
	statusSources := []struct {
		matches    map[discovery.StatusType][]Match
		sourceType string
	}{{s.Metrics, "metrics"}, {s.Statements, "statements"}}
	for _, statusSource := range statusSources {
		for statusType, statements := range statusSource.matches {
			if ok, e := discovery.IsValidStatus(statusType); !ok {
				err = multierr.Combine(err, e)
				continue
			}
			for _, logMatch := range statements {
				var matchTypes []string
				if logMatch.Strict != "" {
					matchTypes = append(matchTypes, "strict")
				}
				if logMatch.Regexp != "" {
					matchTypes = append(matchTypes, "regexp")
				}
				if logMatch.Expr != "" {
					matchTypes = append(matchTypes, "expr")
				}
				if len(matchTypes) != 1 {
					err = multierr.Combine(err, fmt.Errorf(
						"`%s` status source type `%s` match type validation failed. Must provide one of %v but received %v", statusSource.sourceType, statusType, allowedMatchTypes, matchTypes,
					))
				}
				if e := logMatch.Record.validate(); e != nil {
					err = multierr.Combine(err, fmt.Errorf(" %q log record validation failure: %w", statusType, e))
				}
			}
		}
	}
	return err
}

func (lr *LogRecord) validate() error {
	// TODO: supported severity text validation
	return nil
}

// receiverCreatorFactoryAndConfig will embed the applicable receiver creator fields in a new receiver creator config
// suitable for being used to create a receiver instance by the returned factory.
func (cfg *Config) receiverCreatorFactoryAndConfig(correlations correlationStore) (component.ReceiverFactory, config.Receiver, error) {
	receiverCreatorFactory := receivercreator.NewFactory()
	receiverCreatorDefaultConfig := receiverCreatorFactory.CreateDefaultConfig()
	receiverCreatorConfig, ok := receiverCreatorDefaultConfig.(*receivercreator.Config)
	if !ok {
		return nil, nil, fmt.Errorf("failed to coerce to receivercreator.Config")
	}
	receiverCreatorConfig.SetIDName(cfg.ID().String())
	receiverCreatorConfig.WatchObservers = cfg.WatchObservers

	receiversConfig, err := cfg.receiverCreatorReceiversConfig(correlations)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to produce receiver creator receivers config: %w", err)
	}
	receiverTemplates := confmap.NewFromStringMap(map[string]any{"receivers": receiversConfig})
	if err := receiverCreatorConfig.Unmarshal(receiverTemplates); err != nil {
		return nil, nil, fmt.Errorf("failed unmarshaling discoveryreceiver receiverTemplates into receiver_creator config: %w", err)
	}

	return receiverCreatorFactory, receiverCreatorConfig, nil
}

// receiverCreatorReceiversConfig produces the actual config string map used by the receiver creator config unmarshaler.
func (cfg *Config) receiverCreatorReceiversConfig(correlations correlationStore) (map[string]any, error) {
	receiversConfig := map[string]any{}
	for receiverID, rEntry := range cfg.Receivers {
		resourceAttributes := map[string]string{}
		for k, v := range rEntry.ResourceAttributes {
			resourceAttributes[k] = v
		}
		resourceAttributes[discovery.ReceiverNameAttr] = receiverID.Name()
		resourceAttributes[discovery.ReceiverTypeAttr] = string(receiverID.Type())
		resourceAttributes[receiverRuleAttr] = rEntry.Rule
		resourceAttributes[discovery.EndpointIDAttr] = "`id`"

		if cfg.EmbedReceiverConfig {
			embeddedConfig := map[string]any{}
			embeddedReceiversConfig := map[string]any{}
			receiverConfig := map[string]any{}
			receiverConfig["rule"] = rEntry.Rule
			receiverConfig["config"] = rEntry.Config
			receiverConfig["resource_attributes"] = rEntry.ResourceAttributes
			embeddedReceiversConfig[receiverID.String()] = receiverConfig
			embeddedConfig["receivers"] = embeddedReceiversConfig

			// we don't embed the `watch_observers` array here since it is added
			// on statement or metric evaluator matches by looking up the
			// Endpoint.ID to the originating observer ComponentID
			var configYaml []byte
			var err error
			if configYaml, err = yaml.Marshal(embeddedConfig); err != nil {
				return nil, fmt.Errorf("failed embedding %q receiver config: %w", receiverID.String(), err)
			}
			encoded := base64.StdEncoding.EncodeToString(configYaml)
			resourceAttributes[discovery.ReceiverConfigAttr] = encoded
			correlations.UpdateAttrs(receiverID, map[string]string{discovery.ReceiverConfigAttr: encoded})
		}

		rEntryMap := map[string]any{}
		rEntryMap["rule"] = rEntry.Rule
		rEntryMap["config"] = rEntry.Config
		rEntryMap["resource_attributes"] = resourceAttributes
		receiversConfig[receiverID.String()] = rEntryMap
	}

	return receiversConfig, nil
}
