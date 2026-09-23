// Copyright Splunk, Inc.
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

// Package parity is a black-box parity test framework for UF-compatible agents.
//
// It runs the same inputs through a Splunk Universal Forwarder (the oracle) and
// any UF-compatible agent (the candidate). Both forward into one real Splunk
// (the Backend) with their own native transports, and the two are compared as
// the Splunk events that land there, read back by search. Agents are never a
// selectable matrix axis: a run is always oracle vs candidate.
//
// The core is dependency-light on purpose (stdlib + yaml for case loading). The
// Docker/testcontainers dependency lives only in a Backend implementation, so
// importing the core alone stays cheap.
package parity

import (
	"context"
	"net"
	"strconv"
	"time"
)

// Record is one indexed Splunk event, read back from a search against the
// backend. It holds only the fields parity is defined on, named as Splunk
// surfaces them in search results:
//
//	Raw        <- _raw   (event text)
//	Time       <- _time
//	Host       <- host
//	Source     <- source
//	Sourcetype <- sourcetype
//	Index      <- index
//	Fields     <- any other search fields (punct, linecount, date_*, custom)
//
// Fields values are strings for now; typed values (numeric index-time fields)
// are a follow-up on the structured-processing axis.
type Record struct {
	Time       time.Time
	Fields     map[string]string
	Raw        string
	Host       string
	Source     string
	Sourcetype string
	Index      string
}

// Endpoint is a host:port an agent forwards to.
type Endpoint struct {
	Host string
	Port int
}

func (e Endpoint) String() string {
	return net.JoinHostPort(e.Host, strconv.Itoa(e.Port))
}

// HEC is the HTTP Event Collector coordinate a Backend exposes for agents to
// forward into. The same endpoint serves every HEC route: the collector's
// splunkhecexporter posts JSON to /services/collector/event, and UF httpout
// posts cooked S2S to /services/collector/s2s. Clients append the route their
// transport needs.
type HEC struct {
	Endpoint string // e.g. https://127.0.0.1:32769
	Token    string // HEC token (UUID form; UF httpout requires >= 36 chars)
}

// Backend is the shared Splunk instance both agents forward into. It replaces
// the earlier wire-decoding sink: rather than decode a transport, each agent
// forwards with its own native transport and the backend validates by querying
// indexed events over the REST search API. Because both agents land in the same
// real Splunk, comparison is true UF-vs-candidate parity, not agent-vs-golden.
//
// Agents are kept apart by index (Search/Clean take an SPL query that scopes to
// one agent's index), so a run needs no clean-between-agents step.
type Backend interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	// HEC returns the endpoint and token agents forward to.
	HEC() HEC
	// Search runs an SPL query and returns matching events as Records.
	Search(ctx context.Context, spl string) ([]Record, error)
	// Clean removes previously indexed events matching spl.
	Clean(ctx context.Context, spl string) error
}

// Adapter drives one agent through a case, hiding config layout and lifecycle so
// a single test case runs unchanged on every agent. Name is a report label, not
// a selector.
type Adapter interface {
	Name() string
	// InstallDir is the agent root, used to interpolate AGENT_DIR.
	InstallDir() string
	// Prepare installs the case's already-interpolated config files from
	// configDir (inputs.conf, outputs.conf) into the agent's own config layout,
	// arranging to restore prior state on Cleanup.
	Prepare(configDir string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup() error
}

// Validator compares a reference capture (a live UF run, or a recorded golden /
// hand-authored expectation) against a candidate capture. It diffs the fields
// the reference sets and ignores everything else, so volatile keys (indextime,
// stream/ACK ids) are excluded simply by never appearing in the reference.
type Validator interface {
	Validate(reference, candidate []Record) Result
}

// Result is the outcome of a comparison.
type Result struct {
	Mismatches []Mismatch
	Match      bool
}

// Mismatch is one field-level difference for reporting.
type Mismatch struct {
	Field    string
	Expected string
	Actual   string
	Record   int
}
