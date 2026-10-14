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

package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFamily struct {
	verbs   []VerbInfo
	calls   []call
	wantErr bool
}

type call struct {
	verb string
	args []string
}

func (f *fakeFamily) Verbs() []VerbInfo { return f.verbs }

func (f *fakeFamily) Dispatch(verb string, args []string, _ IO) int {
	f.calls = append(f.calls, call{verb: verb, args: args})
	if f.wantErr {
		return ExitFailure
	}
	return ExitOK
}

func verbInfos(names ...string) []VerbInfo {
	infos := make([]VerbInfo, len(names))
	for i, n := range names {
		infos[i] = VerbInfo{Name: n, Help: n + " help text"}
	}
	return infos
}

func TestDispatch(t *testing.T) {
	tests := map[string]struct {
		wantVerb string
		args     []string
		wantRest []string
		wantOK   bool
	}{
		"empty args": {
			args:   []string{},
			wantOK: false,
		},
		"nil args": {
			args:   nil,
			wantOK: false,
		},
		"bare recognized verb": {
			args:     []string{"start"},
			wantOK:   true,
			wantVerb: "start",
			wantRest: []string{},
		},
		"verb with trailing args": {
			args:     []string{"start", "--config", "/x"},
			wantOK:   true,
			wantVerb: "start",
			wantRest: []string{"--config", "/x"},
		},
		"single dash flag (not help)": {
			args:   []string{"-x"},
			wantOK: false,
		},
		"double dash flag": {
			args:   []string{"--config", "/x"},
			wantOK: false,
		},
		"otelcol_options shape": {
			args:   []string{"--config=/etc/otel/collector/agent_config.yaml", "--discovery"},
			wantOK: false,
		},
		"unknown bare token": {
			args:   []string{"validate"},
			wantOK: false,
		},
		"verb-like but cased": {
			args:   []string{"Start"},
			wantOK: false,
		},
		"verb only honored at args[0]": {
			args:   []string{"--config", "x", "start"},
			wantOK: false,
		},
		"empty first arg": {
			args:   []string{""},
			wantOK: false,
		},
		"bare dash": {
			args:   []string{"-"},
			wantOK: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := &fakeFamily{verbs: verbInfos("start", "stop", "status", "restart")}
			code, ok := Dispatch([]Family{f}, tt.args, IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})

			assert.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				assert.Empty(t, f.calls)
				return
			}
			require.Len(t, f.calls, 1)
			assert.Equal(t, tt.wantVerb, f.calls[0].verb)
			assert.Equal(t, tt.wantRest, f.calls[0].args)
			assert.Equal(t, ExitOK, code)
		})
	}
}

func TestDispatchPropagatesFamilyExitCode(t *testing.T) {
	f := &fakeFamily{verbs: verbInfos("start"), wantErr: true}
	code, ok := Dispatch([]Family{f}, []string{"start"}, IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})
	assert.True(t, ok)
	assert.Equal(t, ExitFailure, code)
}

func TestDispatchFirstRegisteredFamilyWins(t *testing.T) {
	first := &fakeFamily{verbs: verbInfos("start")}
	second := &fakeFamily{verbs: verbInfos("start")}

	_, ok := Dispatch([]Family{first, second}, []string{"start"}, IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})

	assert.True(t, ok)
	assert.Len(t, first.calls, 1)
	assert.Empty(t, second.calls)
}

func TestDispatchChecksEachFamilyInOrder(t *testing.T) {
	lifecycleFamily := &fakeFamily{verbs: verbInfos("start", "stop", "status", "restart")}
	otherFamily := &fakeFamily{verbs: verbInfos("reload")}

	code, ok := Dispatch([]Family{lifecycleFamily, otherFamily}, []string{"reload"}, IO{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}})

	assert.True(t, ok)
	assert.Equal(t, ExitOK, code)
	assert.Empty(t, lifecycleFamily.calls)
	require.Len(t, otherFamily.calls, 1)
	assert.Equal(t, "reload", otherFamily.calls[0].verb)
}

func TestDispatchHelpVerbs(t *testing.T) {
	for _, helpArg := range []string{"help", "-h", "--help"} {
		t.Run(helpArg, func(t *testing.T) {
			f := &fakeFamily{verbs: verbInfos("start", "stop")}
			outBuf := &bytes.Buffer{}
			streams := IO{Out: outBuf, Err: &bytes.Buffer{}}

			code, ok := Dispatch([]Family{f}, []string{helpArg}, streams)

			assert.True(t, ok)
			assert.Equal(t, ExitOK, code)
			assert.Empty(t, f.calls, "help must not dispatch to any family")
			assert.Contains(t, outBuf.String(), "start")
			assert.Contains(t, outBuf.String(), "stop")
		})
	}
}

func TestDispatchHelpWithTrailingArgsIgnored(t *testing.T) {
	f := &fakeFamily{verbs: verbInfos("start")}
	outBuf := &bytes.Buffer{}

	code, ok := Dispatch([]Family{f}, []string{"help", "extra", "args"}, IO{Out: outBuf, Err: &bytes.Buffer{}})

	assert.True(t, ok)
	assert.Equal(t, ExitOK, code)
	assert.NotEmpty(t, outBuf.String())
}

func TestHelpText(t *testing.T) {
	families := []Family{
		&fakeFamily{verbs: verbInfos("start", "stop")},
		&fakeFamily{verbs: verbInfos("reload")},
	}

	text := HelpText(families)

	assert.Contains(t, text, "Usage: otelcollauncher")
	assert.Contains(t, text, "start")
	assert.Contains(t, text, "start help text")
	assert.Contains(t, text, "stop")
	assert.Contains(t, text, "reload")
}

func TestHelpTextNoFamilies(t *testing.T) {
	text := HelpText(nil)
	assert.Contains(t, text, "Usage: otelcollauncher")
	assert.NotContains(t, text, "Commands:")
}
