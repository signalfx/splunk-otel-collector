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
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
	"unicode/utf16"
)

func utf16Bytes(text string, bigEndian bool) []byte {
	out := []byte{0xff, 0xfe}
	if bigEndian {
		out = []byte{0xfe, 0xff}
	}
	for _, u := range utf16.Encode([]rune(text)) {
		if bigEndian {
			out = append(out, byte(u>>8), byte(u))
		} else {
			out = append(out, byte(u), byte(u>>8))
		}
	}
	return out
}

func utf32Bytes(text string, bom []byte, order binary.AppendByteOrder) []byte {
	out := append([]byte{}, bom...)
	for _, r := range text {
		out = order.AppendUint32(out, uint32(r))
	}
	return out
}

func TestParseConfDecodesUTF16AndStripsNULs(t *testing.T) {
	const text = "[s]\r\nk = v\u00e9\U0001F600\r\n"
	for name, tc := range map[string]struct {
		value   string
		payload []byte
	}{
		"utf16le": {payload: utf16Bytes(text, false), value: "v\u00e9\U0001F600"},
		"utf16be": {payload: utf16Bytes(text, true), value: "v\u00e9\U0001F600"},
		"nuls":    {payload: []byte("[s]\nk\x00 = v\x00al\n"), value: "val"},
	} {
		t.Run(name, func(t *testing.T) {
			want := []setting{{stanza: "s", header: true}, {stanza: "s", key: "k", value: tc.value}}
			got, err := parseConf(tc.payload)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("parseConf() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestParseConfRejectsUTF32(t *testing.T) {
	const text = "[s]\nk = v\u00e9\U0001F600\n"
	leBOM := []byte{0xff, 0xfe, 0, 0}
	beBOM := []byte{0, 0, 0xfe, 0xff}
	for name, payload := range map[string][]byte{
		"utf32le":       utf32Bytes(text, leBOM, binary.LittleEndian),
		"utf32be":       utf32Bytes(text, beBOM, binary.BigEndian),
		"utf32le empty": leBOM,
		"utf32be empty": beBOM,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := parseConf(payload)
			if !errors.Is(err, errUnsupportedEncoding) {
				t.Fatalf("parseConf() error = %v, want %v", err, errUnsupportedEncoding)
			}
			if got != nil {
				t.Errorf("parseConf() = %#v, want no settings", got)
			}
		})
	}
}

func TestParseConfBOMLookalikesAreNotUTF32(t *testing.T) {
	// Only the complete four-byte UTF-32 BOMs are rejected; an empty UTF-16
	// file and a UTF-16 file whose text begins with a non-NUL are still valid.
	for name, payload := range map[string][]byte{
		"utf16le empty": {0xff, 0xfe},
		"utf16be empty": {0xfe, 0xff},
		"utf16le short": {0xff, 0xfe, 'a'},
		"leading NULs":  {0, 0, 'k', '=', 'v'},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseConf(payload); err != nil {
				t.Fatalf("parseConf() error = %v", err)
			}
		})
	}
}

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
			got, err := parseConf([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
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
		"$SPLUNK_HOME_DIR":               "$SPLUNK_HOME_DIR",
		"$$SPLUNK_HOME":                  "$/splunk",
		"cost $5 $":                      "cost $5 $",
	}
	for input, want := range tests {
		if got := expandStanzaName(input, vars); got != want {
			t.Errorf("expandStanzaName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExpandStanzaNameWithWindowsPaths(t *testing.T) {
	vars := expansionVars{
		home:   `C:\SplunkUniversalForwarder`,
		etcDir: `D:\SplunkConfig`,
		lookupEnv: func(string) (string, bool) {
			return "", false
		},
	}
	tests := map[string]string{
		`$SPLUNK_HOME\etc`:           `D:\SplunkConfig`,
		`$SPLUNK_HOME\etc\apps`:      `D:\SplunkConfig\apps`,
		`$SPLUNK_HOME/etc\apps`:      `D:\SplunkConfig\apps`,
		`$SPLUNK_HOME\etc/apps`:      `D:\SplunkConfig/apps`,
		`$SPLUNK_HOME\etcx`:          `C:\SplunkUniversalForwarder\etcx`,
		`$SPLUNK_HOME\etc-other`:     `C:\SplunkUniversalForwarder\etc-other`,
		`$SPLUNK_ETC\apps`:           `D:\SplunkConfig\apps`,
		`monitor://$SPLUNK_HOME\var`: `monitor://C:\SplunkUniversalForwarder\var`,
	}
	for input, want := range tests {
		if got := expandStanzaName(input, vars); got != want {
			t.Errorf("expandStanzaName(%q) = %q, want %q", input, got, want)
		}
	}
}
