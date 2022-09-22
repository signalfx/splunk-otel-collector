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
	"sync"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	observerIDAttr = "discovery.observer.id"
)

// evaluator is the base status matcher that determines if telemetry warrants emitting a matching log record.
// It also provides embedded config correlation that its embedding structs will utilize.
type evaluator struct {
	logger       *zap.Logger
	config       *Config
	correlations correlationStore
	// if match.FirstOnly this ~sync.Map(map[string]struct{}) keeps track of
	// whether we've already emitted a record for the statement and can skip processing.
	alreadyLogged *sync.Map
	exprEnv       func(pattern string) map[string]any
}

// evaluateMatch parses the provided Match and returns whether it warrants a status log record
func (e *evaluator) evaluateMatch(match Match, pattern, status string, receiverID config.ComponentID, endpointID observer.EndpointID) (bool, error) {
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
		// TODO: cache compiled programs for performance benefit
		if program, err = expr.Compile(match.Expr, expr.Env(e.exprEnv(pattern))); err != nil {
			err = fmt.Errorf("invalid match expr statement: %w", err)
		} else {
			matchFunc = func(p string) (bool, error) {
				ret, runErr := vm.Run(program, e.exprEnv(p))
				if runErr != nil {
					return false, runErr
				}
				return ret.(bool), nil
			}
		}
	default:
		err = fmt.Errorf("no valid match field provided")
	}
	if err != nil {
		return false, err
	}

	var shouldLog bool
	shouldLog, err = matchFunc(pattern)
	if !shouldLog || err != nil {
		return false, err
	}

	if match.FirstOnly {
		loggedKey := fmt.Sprintf("%s::%s::%s::%s", endpointID, receiverID.String(), status, matchPattern)
		if _, ok := e.alreadyLogged.LoadOrStore(loggedKey, struct{}{}); ok {
			shouldLog = false
		}
	}

	return shouldLog, nil
}

// correlateResourceAttributes will copy all `from` attributes to `to` in addition to
// updating embedded base64 config content, if configured, to include the correlated observer ID
// that is otherwise unavailable to status sources.
func (e *evaluator) correlateResourceAttributes(from, to pcommon.Map, corr correlation) {
	receiverType := string(corr.receiverID.Type())
	receiverName := corr.receiverID.Name()

	observerID := corr.observerID.String()
	to.PutString(observerIDAttr, observerID)

	var receiverAttrs map[string]string
	hasTemporaryReceiverConfigAttr := false
	receiverAttrs = e.correlations.Attrs(corr.receiverID)

	if e.config.EmbedReceiverConfig {
		if _, ok := from.Get(receiverConfigAttr); !ok {
			// statements don't inherit embedded configs in their resource attributes
			// from the receiver creator, so we should temporarily include it in `from`
			// so as not to mutate the original while providing the desired receiver config
			// value set by the initial receiver config parser.
			from.PutString(receiverConfigAttr, receiverAttrs[receiverConfigAttr])
			hasTemporaryReceiverConfigAttr = true
		}
	}

	from.Range(func(k string, v pcommon.Value) bool {
		if k == receiverConfigAttr && e.config.EmbedReceiverConfig {
			configVal := v.AsString()
			if updatedConfig, ok := receiverAttrs[receiverUpdatedConfigAttr]; ok {
				configVal = updatedConfig
			} else {
				var err error
				if updatedConfig, err = addObserverToEncodedConfig(configVal, observerID); err != nil {
					// log failure and continue with existing config sans observer
					e.logger.Debug(fmt.Sprintf("failed adding %q to %s", observerID, receiverConfigAttr), zap.String("receiver.type", receiverType), zap.String("receiver.name", receiverName), zap.Error(err))
				} else {
					e.logger.Debug("Adding watch_observer to embedded receiver config receiver attrs", zap.String("observer", corr.observerID.String()), zap.String("receiver.type", receiverType), zap.String("receiver.name", receiverName))
					e.correlations.UpdateAttrs(corr.receiverID, map[string]string{
						receiverUpdatedConfigAttr: updatedConfig,
					})
					configVal = updatedConfig
				}
			}
			v = pcommon.NewValueString(configVal)
		}
		to.PutString(k, v.AsString())
		return true
	})
	if hasTemporaryReceiverConfigAttr {
		from.Remove(receiverConfigAttr)
	}
}

func addObserverToEncodedConfig(encoded, observerID string) (string, error) {
	cfg := map[string]any{}
	dBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if err = yaml.Unmarshal(dBytes, &cfg); err != nil {
		return "", err
	}
	cfg["watch_observers"] = []string{observerID}

	var cfgYaml []byte
	if cfgYaml, err = yaml.Marshal(cfg); err != nil {
		return "", fmt.Errorf("failed embedding receiver config to include %q: %w", observerID, err)
	}
	return base64.StdEncoding.EncodeToString(cfgYaml), nil
}
