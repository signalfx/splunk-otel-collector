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
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"github.com/signalfx/splunk-otel-collector/internal/common/discovery"
)

var (
	endpointIDRegexp = regexp.MustCompile(`^.*{endpoint=.*}/(?P<id>.*)$`)
	undesiredFields  = []string{"ts", "msg", "level"}
)

// Statement models a zapcore.Entry but defined here for usability/maintainability
type Statement struct {
	Message    string
	Fields     map[string]any
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

// ReceiverNameToIDs parses the zap "name" field value according to
// outcome of https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12670
// where receiver creator receiver names are of the form
// `<receiver.type>/<receiver.name>/receiver_creator/<receiver-creator.name>{endpoint="<Endpoint.Target>"}/<Endpoint.ID>`.
// If receiverName argument is not of this form empty Component and Endpoint IDs are returned.
func ReceiverNameToIDs(statement *Statement) (component.ID, observer.EndpointID) {
	// The receiver creator sets dynamically created receiver names as the zap "name" field for their component logger.
	nameField, ok := statement.Fields["name"]
	if !ok {
		// there is nothing we can do without a name field
		return discovery.NoType, ""
	}

	// receiver creator generated message names must contain one "/receiver_creator/"
	nameFieldParts := strings.Split(fmt.Sprintf("%s", nameField), "/receiver_creator/")
	if len(nameFieldParts) != 2 {
		// invalid format of the name field
		return discovery.NoType, ""
	}
	receiverIDSection := nameFieldParts[0]
	endpointSection := nameFieldParts[1]

	receiverIDParts := strings.SplitN(receiverIDSection, "/", 2)
	receiverType, err := component.NewType(receiverIDParts[0])
	if err != nil {
		// receiver type is invalid
		return discovery.NoType, ""
	}
	var receiverName string
	if len(receiverIDParts) > 1 {
		receiverName = receiverIDParts[1]
	}

	endpointMatches := endpointIDRegexp.FindStringSubmatch(endpointSection)
	if len(endpointMatches) < 1 {
		// endpoint ID is not present
		return discovery.NoType, ""
	}
	endpointID := observer.EndpointID(endpointMatches[1])

	return component.MustNewIDWithName(receiverType.String(), receiverName), endpointID
}
