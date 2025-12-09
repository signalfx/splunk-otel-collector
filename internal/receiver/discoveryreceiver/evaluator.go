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
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

// exprEnvFunc is to create an expr.Env function from pattern content.
type exprEnvFunc func(pattern string) map[string]any

// evaluator is the base status matcher that determines if telemetry warrants emitting a matching log record.
// It also provides embedded config correlation that its embedding structs will utilize.
type evaluator struct {
	logger       *zap.Logger
	config       *Config
	correlations *correlationStore
	// this ~sync.Map(map[string]struct{}) keeps track of
	// whether we've already emitted a record for the statement and can skip processing.
	alreadyLogged *sync.Map
	exprEnv       exprEnvFunc
}

func newEvaluator(logger *zap.Logger, config *Config, correlations *correlationStore, envFunc exprEnvFunc) *evaluator {
	return &evaluator{
		logger:        logger,
		config:        config,
		correlations:  correlations,
		alreadyLogged: &sync.Map{},
		exprEnv:       envFunc,
	}
}

// evaluateMatch parses the provided Match and returns whether it warrants a status log record
func (e *evaluator) evaluateMatch(match Match, pattern string, status discovery.StatusType, receiverID component.ID, endpointID observer.EndpointID) (bool, error) {
	var matchFunc func(p string) (bool, error)
	var matchPattern string

	var err error
	switch {
	case match.Strict != "":
		matchPattern = match.Strict
		matchFunc = func(p string) (bool, error) {
			return p == match.Strict, nil
		}
	case match.Regexp != "":
		matchPattern = match.Regexp
		var re *regexp.Regexp
		if re, err = regexp.Compile(matchPattern); err != nil {
			err = fmt.Errorf("invalid match regexp statement: %w", err)
		} else {
			matchFunc = func(p string) (bool, error) { return re.MatchString(p), nil }
		}
	case match.Expr != "":
		matchPattern = match.Expr
		var program *vm.Program
		// we need a way to look up fields that aren't valid identifiers https://github.com/antonmedv/expr/issues/106
		env := e.exprEnv(pattern)
		env["ExprEnv"] = env
		// TODO: cache compiled programs for performance benefit
		if program, err = expr.Compile(match.Expr, expr.Env(env)); err != nil {
			err = fmt.Errorf("invalid match expr statement: %w", err)
		} else {
			matchFunc = func(_ string) (bool, error) {
				ret, runErr := vm.Run(program, env)
				if runErr != nil {
					return false, runErr
				}
				return ret.(bool), nil
			}
		}
	default:
		err = errors.New("no valid match field provided")
	}
	if err != nil {
		return false, err
	}

	var shouldLog bool
	shouldLog, err = matchFunc(pattern)
	if !shouldLog || err != nil {
		return false, err
	}

	loggedKey := fmt.Sprintf("%s::%s::%s::%s", endpointID, receiverID.String(), status, matchPattern)
	if _, ok := e.alreadyLogged.LoadOrStore(loggedKey, struct{}{}); ok {
		shouldLog = false
	}

	e.logger.Debug(fmt.Sprintf("evaluated match %v against %q (should log: %v)", matchPattern, pattern, shouldLog))
	return shouldLog, nil
}

// correlateResourceAttributes sets correlation attributes including embedded base64 config content, if configured.
func (e *evaluator) correlateResourceAttributes(cfg *Config, to map[string]string, corr correlation) {
	observerID := corr.observerID.String()
	if observerID != "" && observerID != discovery.NoType.String() {
		to[discovery.ObserverIDAttr] = observerID
	}

	rEntry := cfg.Receivers[corr.receiverID] // it's safe to assume this exists.
	if meta, exists := receiverMetaMap[corr.receiverID.String()]; exists {
		to[serviceTypeAttr] = meta.ServiceType
	}

	if e.config.EmbedReceiverConfig {
		embeddedConfig := map[string]any{}
		embeddedReceiversConfig := map[string]any{}
		receiverConfig := map[string]any{}
		receiverConfig["rule"] = rEntry.Rule
		receiverConfig["config"] = rEntry.Config
		receiverConfig["resource_attributes"] = rEntry.ResourceAttributes
		embeddedReceiversConfig[corr.receiverID.String()] = receiverConfig
		embeddedConfig["receivers"] = embeddedReceiversConfig
		if observerID != "" && observerID != discovery.NoType.String() {
			embeddedConfig["watch_observers"] = []string{observerID}
		}
		cfgYaml, err := yaml.Marshal(embeddedConfig)
		if err != nil {
			e.logger.Error("failed embedding receiver config", zap.String("observer", observerID), zap.Error(err))
		}
		to[discovery.ReceiverConfigAttr] = base64.StdEncoding.EncodeToString(cfgYaml)
	}
}
