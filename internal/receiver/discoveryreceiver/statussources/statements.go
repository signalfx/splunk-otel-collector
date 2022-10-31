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

package statussources

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

var (
	ReceiverCreatorRegexp = regexp.MustCompile(`receiver_creator/`)
	receiverNameRegexp    = regexp.MustCompile(`^(?P<type>[^/]+)/(?P<name>.*)$`)
	EndpointTargetRegexp  = regexp.MustCompile(`{endpoint=[^}]*}/`)
	endpointIDRegexp      = regexp.MustCompile(`^.*{endpoint=.*}/(?P<id>.*)$`)
	undesiredFields       = []string{"ts", "msg", "level"}
)

// Statement models a zapcore.Entry but defined here for usability/maintainability
type Statement struct {
	Message    string
	Fields     map[string]any
	Level      string
	Time       time.Time
	LoggerName string
	Caller     zapcore.EntryCaller
	Stack      string
}

func NewZapCoreEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
}

// StatementFromZapCoreEntry converts the Entry to a Statement using the provided encoder, which is assumed to be
// a JSONEncoder (unexpected from zapcore) to obtain the target Fields.
func StatementFromZapCoreEntry(encoder zapcore.Encoder, entry zapcore.Entry, fields []zapcore.Field) (*Statement, error) {
	statement := &Statement{
		Message:    entry.Message,
		Level:      entry.Level.String(),
		Time:       entry.Time,
		LoggerName: entry.LoggerName,
		Caller:     entry.Caller,
		Stack:      entry.Stack,
	}
	var err error
	var entryBuffer *buffer.Buffer

	if entryBuffer, err = encoder.EncodeEntry(entry, fields); err != nil {
		return nil, fmt.Errorf("failed encoding zapcore.Entry: %w", err)
	}

	b := entryBuffer.Bytes()
	if err = json.Unmarshal(b, &statement.Fields); err != nil {
		return nil, fmt.Errorf("failed representing encoded zapcore.Entry (%s) as json: %w", b, err)
	}

	for _, undesiredField := range undesiredFields {
		delete(statement.Fields, undesiredField)
	}

	return statement, nil
}

func (s *Statement) ToLogRecord() plog.LogRecord {
	logRecord := plog.NewLogRecord()
	if s == nil {
		return logRecord
	}

	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(s.Time))
	logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	logRecord.Body().SetStr(s.Message)
	logRecord.SetSeverityText(s.Level)
	logRecord.Attributes().FromRaw(s.Fields)
	return logRecord
}

// ReceiverNameToIDs parses the zap "name" field value according to
// outcome of https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12670
// where receiver creator receiver names are of the form
// `<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`.
// If receiverName argument is not of this form empty Component and Endpoint IDs are returned.
func ReceiverNameToIDs(record plog.LogRecord) (receiverID config.ComponentID, endpointID observer.EndpointID) {
	// The receiver creator sets dynamically created receiver names as the zap "name" field for their component logger.
	nameAttr, ok := record.Attributes().Get("name")
	if !ok {
		// there is nothing we can do without a receiver name
		return discovery.NoType, ""
	}
	receiverName := nameAttr.AsString()

	// The receiver creator will log an initial start statement not from the underlying receiver's logger.
	// These statements have an "endpoint_id" field and the "name" field won't include the necessary "receiver_creator/"
	// and "{endpoint=<endpoint.Target>}" separators. In this case we get the EndpointID, if any, and form a placeholder name of desired form.
	if endpointIDAttr, hasEndpointID := record.Attributes().Get("endpoint_id"); hasEndpointID {
		receiverName = fmt.Sprintf(`%s/receiver_creator/<PLACEHOLDER>/{endpoint="PLACEHOLDER"}/%s`, receiverName, endpointIDAttr.AsString())
	}

	// receiver creator generated and altered initial endpoint handler message names must contain
	// one "receiver_creator" and one "{endpoint=<Endpoint.Target>}" separator or are unable to be decomposed
	for _, re := range []*regexp.Regexp{ReceiverCreatorRegexp, EndpointTargetRegexp} {
		if matches := re.FindAllStringSubmatch(receiverName, -1); len(matches) != 1 {
			return discovery.NoType, ""
		}
	}

	var rcIdx int
	if rcIdx = strings.Index(receiverName, "receiver_creator/"); rcIdx == -1 {
		// previous check enforces this to not be the case but for good measure
		return discovery.NoType, ""
	}
	nameSection := receiverName[:rcIdx]
	endpointSection := receiverName[rcIdx:]

	var nameMatches []string
	if nameMatches = receiverNameRegexp.FindStringSubmatch(nameSection); len(nameMatches) < 2 {
		return discovery.NoType, ""
	}
	rType := nameMatches[1]

	var nameCandidate string
	if len(nameMatches) > 2 {
		nameCandidate = nameMatches[2]
	}
	var rName string
	if nameCandidate != "" {
		rName = nameCandidate
		if nameCandidate[len(nameCandidate)-1] == '/' {
			rName = nameCandidate[0 : len(nameCandidate)-1]
		}
	}
	var eID string
	if endpointMatches := endpointIDRegexp.FindStringSubmatch(endpointSection); len(endpointMatches) > 1 {
		eID = endpointMatches[1]
	}
	return config.NewComponentIDWithName(config.Type(rType), rName), observer.EndpointID(eID)
}
