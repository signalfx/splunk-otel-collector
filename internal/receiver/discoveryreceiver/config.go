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

	"go.opentelemetry.io/collector/config"
	"go.uber.org/multierr"
)

var (
	_ config.Receiver = (*Config)(nil)

	allowedStatusTypes = []string{"successful", "partial", "failed"}
	allowedStatuses    = func() map[string]struct{} {
		sm := map[string]struct{}{}
		for _, status := range allowedStatusTypes {
			sm[status] = struct{}{}
		}
		return sm
	}()

	allowedMatchTypes = []string{"regexp", "strict", "expr"}
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
	Metrics    map[string][]Match `mapstructure:"metrics"`
	Statements map[string][]Match `mapstructure:"statements"`
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
		if e := rEntry.validate(); e != nil {
			err = multierr.Combine(err, fmt.Errorf("receiver %q validation failure: %w", rName, e))
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
		return fmt.Errorf("`status` must contain at least one `metrics` or `statements` mapping with at least one of %v", allowedStatusTypes)
	}

	var err error
	statusSources := []struct {
		matches    map[string][]Match
		sourceType string
	}{{s.Metrics, "metrics"}, {s.Statements, "statements"}}
	for _, statusSource := range statusSources {
		for statusType, statements := range statusSource.matches {
			if _, ok := allowedStatuses[statusType]; !ok {
				err = multierr.Combine(err, fmt.Errorf("unsupported status %q. must be one of %v", statusType, allowedStatusTypes))
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
