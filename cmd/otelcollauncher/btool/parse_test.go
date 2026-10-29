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

package btool

import (
	"reflect"
	"testing"
)

func TestParseConf(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []setting
	}{
		{
			name:  "comments only at line start",
			input: "[s]\n# c = 1\n  ; d = 2\na = x # y\nb = x;y\n",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: "a", value: "x # y"}, {stanza: "s", key: "b", value: "x;y"}},
		},
		{
			name:  "values are not unquoted and may contain delimiters",
			input: "[s]\na = \"q\"\nb = c:d=e\n",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: "a", value: `"q"`}, {stanza: "s", key: "b", value: "c:d=e"}},
		},
		{
			name:  "lines without equals are ignored",
			input: "[s]\njustakey\na = 1\n",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: "a", value: "1"}},
		},
		{
			name:  "keys before any stanza belong to default",
			input: "a = 1\n[s]\n",
			want:  []setting{{stanza: "default", key: "a", value: "1"}, {stanza: "s", header: true}},
		},
		{
			name:  "continuation removes one backslash and swallows any next line",
			input: "[s]\na = one\\\n[t]\nb = x\\\\\nc = 2\n",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: "a", value: "one\n[t]"}, {stanza: "s", key: "b", value: "x\\\nc = 2"}},
		},
		{
			name:  "continuation at end of file",
			input: "[s]\na = one\\",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: "a", value: "one"}},
		},
		{
			name:  "keys are unescaped but values are not",
			input: "[s]\na\\=b\\\\c = d\\\\e\n",
			want:  []setting{{stanza: "s", header: true}, {stanza: "s", key: `a=b\c`, value: `d\\e`}},
		},
		{
			name:  "stanza header spans first bracket to last bracket",
			input: "\xef\xbb\xbf[a[b]c] junk\r\nk = v\r\n[open\n[]\n",
			want: []setting{
				{stanza: "a[b]c", header: true},
				{stanza: "a[b]c", key: "k", value: "v"},
				{stanza: "open", header: true},
				{stanza: "", header: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseConf([]byte(tt.input)); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseConf() =\n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}

func TestExpandStanzaName(t *testing.T) {
	env := map[string]string{"CUSTOM": "/custom"}
	vars := expansionVars{
		etcDir: "/splunk/etc",
		home:   "/splunk",
		lookupEnv: func(name string) (string, bool) {
			value, ok := env[name]
			return value, ok
		},
	}
	tests := map[string]string{
		"monitor://$SPLUNK_HOME/var/log": "monitor:///splunk/var/log",
		"$SPLUNK_HOME/etc/apps":          "/splunk/etc/apps",
		"$SPLUNK_HOME/etcx":              "/splunk/etcx",
		"$SPLUNK_ETC":                    "/splunk/etc",
		"$SPLUNK_DB/x":                   "/splunk/var/lib/splunk/x",
		"$CUSTOM/x":                      "/custom/x",
		"${SPLUNK_HOME}":                 "${SPLUNK_HOME}",
		"$UNDEFINED/x":                   "$UNDEFINED/x",
		"$SPLUNK_HOMEx":                  "$SPLUNK_HOMEx",
		"$$SPLUNK_HOME":                  "$/splunk",
		"cost $5 $":                      "cost $5 $",
	}
	for input, want := range tests {
		if got := expandStanzaName(input, vars); got != want {
			t.Errorf("expandStanzaName(%q) = %q, want %q", input, got, want)
		}
	}
}
