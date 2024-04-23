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
	"encoding/json"
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
	"github.com/signalfx/splunk-otel-collector/internal/receiver/discoveryreceiver/statussources"
)

var _ zapcore.Core = (*statementEvaluator)(nil)

const statementMatch = "statement.match"

// statementEvaluator conforms to a zapcore.Core to intercept component log statements and
// determine if they match any configured Status match rules. If so, they emit log records
// for the matching statement.
type statementEvaluator struct {
	*evaluator
	pLogs chan plog.Logs
	// this is the logger to share with other components to evaluate their statements and produce plog.Logs
	evaluatedLogger *zap.Logger
	encoder         zapcore.Encoder
	id              component.ID
}

func newStatementEvaluator(logger *zap.Logger, id component.ID, config *Config, pLogs chan plog.Logs, correlations correlationStore) (*statementEvaluator, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapConfig.Sampling.Initial = 1
	zapConfig.Sampling.Thereafter = 1
	encoder := statussources.NewZapCoreEncoder()

	se := &statementEvaluator{
		pLogs:   pLogs,
		encoder: encoder,
		id:      id,
	}
	se.evaluator = newEvaluator(logger, config, correlations, se.exprEnv)

	var err error
	if se.evaluatedLogger, err = zapConfig.Build(
		// zap.OnFatal must not be WriteThenFatal or WriteThenNoop since it's rewritten to be WriteThenFatal
		// https://github.com/uber-go/zap/blob/e06e09a6d396031c89b87383eef3cad6f647cf2c/logger.go#L315.
		// Using an arbitrary action offset.
		zap.WithFatalHook(zapcore.WriteThenFatal+100),
		zap.WrapCore(func(_ zapcore.Core) zapcore.Core { return se }),
	); err != nil {
		return nil, err
	}
	return se, nil
}

// exprEnv will unpack logged statement message and field content for expr program use
func (se *statementEvaluator) exprEnv(pattern string) map[string]any {
	patternMap := map[string]any{}
	if err := json.Unmarshal([]byte(pattern), &patternMap); err != nil {
		se.logger.Info(fmt.Sprintf("failed unmarshaling pattern map %q", pattern), zap.Error(err))
		patternMap = map[string]any{"message": pattern}
	}
	return patternMap
}

// Enabled is a zapcore.Core method. We should be enabled for all
// levels since we want to intercept all statements.
func (se *statementEvaluator) Enabled(zapcore.Level) bool {
	return true
}

// With is a zapcore.Core method. We clone ourselves so all
// modified downstream loggers are still evaluated.
func (se *statementEvaluator) With(fields []zapcore.Field) zapcore.Core {
	cp := *se
	cp.encoder = se.encoder.Clone()
	for i := range fields {
		fields[i].AddTo(cp.encoder)
	}
	return &cp
}

// Check is a zapcore.Core method. Similar to Enabled() we want to
// return a valid CheckedEntry for all logging attempts to intercept
// all statements.
func (se *statementEvaluator) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, se)
}

// Write is a zapcore.Core method. This is where the logged entry
// is converted to a statussources.Statement, if from a downstream receiver,
// and its content is evaluated for status matches and plog.Logs translation/submission.
func (se *statementEvaluator) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	statement, err := statussources.StatementFromZapCoreEntry(se.encoder, entry, fields)
	if err != nil {
		return err
	}
	if name, ok := statement.Fields["name"]; ok {
		cid := &component.ID{}
		if err := cid.UnmarshalText([]byte(fmt.Sprintf("%v", name))); err == nil {
			if cid.Type() == component.MustNewType("receiver_creator") && cid.Name() == se.id.String() {
				// this is from our internal Receiver Creator and not a generated receiver, so write
				// it to our logger core without submitting the entry for evaluation
				if ce := se.logger.Check(entry.Level, ""); ce != nil {
					// forward to our logger now that we know entry.Level is accepted
					_ = se.logger.Core().Write(entry, fields)
				}
				return nil
			}
		}
	}

	if pLogs := se.evaluateStatement(statement); pLogs.LogRecordCount() > 0 {
		se.pLogs <- pLogs
	}

	return nil
}

// Sync is a zapcore.Core method.
func (se *statementEvaluator) Sync() error {
	return nil
}

// evaluateStatement will convert the provided statussources.Statement into a plog.Logs with a single log record
// if it matches against the first applicable configured ReceiverEntry's status Statement.[]Match
func (se *statementEvaluator) evaluateStatement(statement *statussources.Statement) plog.Logs {
	se.logger.Debug("evaluating statement", zap.Any("statement", statement))

	receiverID, endpointID, rEntry, shouldEvaluate := se.receiverEntryFromStatement(statement)
	if !shouldEvaluate {
		return plog.NewLogs()
	}

	patternMap := map[string]string{"message": statement.Message}
	for k, v := range statement.Fields {
		switch k {
		case "caller", "monitorID", "name", "stacktrace":
		default:
			patternMap[k] = fmt.Sprintf("%v", v)
		}
	}

	var patternMapStr string
	if pm, err := json.Marshal(patternMap); err != nil {
		se.logger.Debug(fmt.Sprintf("failed marshaling pattern map for %q", statement.Message), zap.Error(err))
		// best effort default in marshaling failure cases
		patternMapStr = fmt.Sprintf(`{"message":%q}`, statement.Message)
	} else {
		patternMapStr = string(pm)
	}
	se.logger.Debug("non-strict matches will be evaluated with pattern map", zap.String("map", patternMapStr))

	for _, match := range rEntry.Status.Statements {
		p := patternMapStr
		if match.Strict != "" {
			p = statement.Message
		}
		if shouldLog, err := se.evaluateMatch(match, p, match.Status, receiverID, endpointID); err != nil {
			se.logger.Info("Error evaluating statement match", zap.Error(err))
			continue
		} else if !shouldLog {
			continue
		}

		entityEvents := experimentalmetricmetadata.NewEntityEventsSlice()
		entityEvent := entityEvents.AppendEmpty()
		entityEvent.ID().PutStr(discovery.EndpointIDAttr, string(endpointID))
		entityState := entityEvent.SetEntityState()
		attrs := entityState.Attributes()
		_ = attrs.FromRaw(statement.Fields)
		corr := se.correlations.GetOrCreate(endpointID, receiverID)
		se.correlateResourceAttributes(se.config, attrs, corr)
		attrs.PutStr(discovery.ReceiverTypeAttr, receiverID.Type().String())
		attrs.PutStr(discovery.ReceiverNameAttr, receiverID.Name())
		attrs.PutStr(discovery.MessageAttr, statement.Message)
		attrs.PutStr(eventTypeAttr, statementMatch)
		attrs.PutStr(receiverRuleAttr, rEntry.Rule.String())

		var desiredRecord LogRecord
		if match.Record != nil {
			desiredRecord = *match.Record
		}
		if desiredRecord.Body != "" {
			body := desiredRecord.Body
			if desiredRecord.AppendPattern {
				body = fmt.Sprintf("%s (evaluated %q)", body, p)
			}
			entityState.Attributes().PutStr(discovery.MessageAttr, body)
		}
		if len(desiredRecord.Attributes) > 0 {
			for k, v := range desiredRecord.Attributes {
				entityState.Attributes().PutStr(k, v)
			}
		}
		entityState.Attributes().PutStr(discovery.StatusAttr, string(match.Status))
		entityEvent.SetTimestamp(pcommon.NewTimestampFromTime(statement.Time))
		return entityEvents.ConvertAndMoveToLogs()
	}
	return plog.NewLogs()
}

func (se *statementEvaluator) receiverEntryFromStatement(statement *statussources.Statement) (component.ID, observer.EndpointID, ReceiverEntry, bool) {
	receiverID, endpointID := statussources.ReceiverNameToIDs(statement)
	if receiverID == discovery.NoType || endpointID == "" {
		// statement evaluation requires both a populated receiver.ID and EndpointID
		se.logger.Debug("unable to evaluate statement from receiver", zap.String("receiver", receiverID.String()))
		return discovery.NoType, "", ReceiverEntry{}, false
	}

	rEntry, ok := se.config.Receivers[receiverID]
	if !ok {
		se.logger.Info("No matching configured receiver for statement status evaluation", zap.String("receiver", receiverID.String()))
		return discovery.NoType, "", ReceiverEntry{}, false
	}

	if rEntry.Status == nil {
		return discovery.NoType, "", ReceiverEntry{}, false
	}

	return receiverID, endpointID, rEntry, true
}
