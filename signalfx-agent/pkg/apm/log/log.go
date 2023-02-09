// Copyright  Splunk, Inc.
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

package log

// Fields is a map that is used to populated logging context.
type Fields map[string]interface{}

type nilLogger struct {
}

func (n nilLogger) Debug(msg string) {
}

func (n nilLogger) Warn(msg string) {
}

func (n nilLogger) Error(msg string) {
}

func (n nilLogger) Info(msg string) {
}

func (n nilLogger) Panic(msg string) {
}

func (n nilLogger) WithFields(fields Fields) Logger {
	return nilLogger{}
}

func (n nilLogger) WithError(err error) Logger {
	return nilLogger{}
}

// Nil logger is a silent logger interface.
var Nil = nilLogger{}

var _ Logger = (*nilLogger)(nil)

// Logger is generic logging interface.
type Logger interface {
	Debug(msg string)
	Warn(msg string)
	Error(msg string)
	Info(msg string)
	Panic(msg string)
	WithFields(fields Fields) Logger
	WithError(err error) Logger
}
