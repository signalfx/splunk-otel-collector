// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package lifecycle implements the otelcollauncher process lifecycle command
// family: start, stop, status, and restart for the collector or supervisor
// process.
package lifecycle

import "github.com/signalfx/splunk-otel-collector/cmd/otelcollauncher/cli"

type verb string

const (
	verbStart   verb = "start"
	verbStop    verb = "stop"
	verbStatus  verb = "status"
	verbRestart verb = "restart"
)

var allVerbs = []cli.VerbInfo{
	{Name: string(verbStart), Help: "start the collector or supervisor process"},
	{Name: string(verbStop), Help: "stop the collector or supervisor process"},
	{Name: string(verbStatus), Help: "report whether the collector or supervisor is running"},
	{Name: string(verbRestart), Help: "restart the collector or supervisor process"},
}

type Manager struct {
	dispatch func(v verb, args []string, streams cli.IO) int
}

func (m *Manager) Verbs() []cli.VerbInfo {
	return allVerbs
}

func (m *Manager) Dispatch(v string, args []string, streams cli.IO) int {
	return m.dispatch(verb(v), args, streams)
}
