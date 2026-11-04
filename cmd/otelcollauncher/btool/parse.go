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
	"errors"
	"strings"
	"unicode/utf16"
)

// defaultStanza is the stanza whose settings every other stanza inherits.
// Splunk matches the name case-sensitively.
const defaultStanza = "default"

// setting is a single key/value pair as it appears in a .conf file. A header
// entry records that a stanza was declared, even if it has no settings.
type setting struct {
	stanza string
	key    string
	value  string
	header bool
}

// parseConf parses the contents of a Splunk .conf file into the ordered list
// of settings it contains. It follows Splunk's lenient rules rather than the
// INI conventions:
//
//   - Only lines starting (after optional whitespace) with # or ; are comments.
//     There are no inline comments, quoting, or alternative delimiters.
//   - Keys are split from values at the first unescaped "=". Lines without one
//     are ignored. In keys, a backslash escapes the next character; values are
//     kept verbatim.
//   - A line ending in a backslash is joined with the next line, whatever that
//     line contains, using a newline; exactly one backslash is removed.
//   - Settings that appear before any stanza belong to the default stanza.
//   - A stanza header is everything between the first "[" and the last "]". A
//     header is recognized on its first physical line, so unlike native btool
//     it cannot be continued with a trailing backslash.
//
// The payload may be UTF-8 (with or without a BOM) or UTF-16 with a BOM. NUL
// characters are removed. A UTF-32 payload, recognized by its BOM, is rejected
// with errUnsupportedEncoding rather than misread.
func parseConf(payload []byte) ([]setting, error) {
	if isUTF32(payload) {
		return nil, errUnsupportedEncoding
	}
	text := strings.TrimPrefix(decodeText(payload), "\xef\xbb\xbf")
	text = strings.ReplaceAll(text, "\x00", "")
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSuffix(line, "\r")
	}

	var settings []setting
	stanza := defaultStanza
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed[0] == '#' || trimmed[0] == ';' {
			continue
		}
		if trimmed[0] == '[' {
			stanza = parseStanzaName(trimmed)
			settings = append(settings, setting{stanza: stanza, header: true})
			continue
		}
		if strings.HasSuffix(line, "\\") {
			var sb strings.Builder
			for {
				if !strings.HasSuffix(line, "\\") {
					sb.WriteString(line)
					break
				}
				sb.WriteString(strings.TrimSuffix(line, "\\"))
				if i+1 >= len(lines) {
					break
				}
				i++
				sb.WriteByte('\n')
				line = lines[i]
			}
			line = sb.String()
		}
		key, value, found := splitSetting(line)
		if !found {
			continue
		}
		settings = append(settings, setting{stanza: stanza, key: key, value: value})
	}
	return settings, nil
}

var errUnsupportedEncoding = errors.New("unsupported UTF-32 encoding; use UTF-8 or UTF-16")

// isUTF32 reports whether the payload starts with a UTF-32 BOM. The
// little-endian one begins with the UTF-16LE BOM, so this must be checked
// before decoding as UTF-16; a UTF-16LE file whose first character is NUL is
// not a meaningful conf file.
func isUTF32(payload []byte) bool {
	return len(payload) >= 4 &&
		(payload[0] == 0xff && payload[1] == 0xfe && payload[2] == 0 && payload[3] == 0 ||
			payload[0] == 0 && payload[1] == 0 && payload[2] == 0xfe && payload[3] == 0xff)
}

// decodeText converts a BOM-prefixed UTF-16 payload to UTF-8 and returns
// anything else unchanged.
func decodeText(payload []byte) string {
	var order func([]byte) uint16
	switch {
	case len(payload) >= 2 && payload[0] == 0xff && payload[1] == 0xfe:
		order = func(b []byte) uint16 { return uint16(b[0]) | uint16(b[1])<<8 }
	case len(payload) >= 2 && payload[0] == 0xfe && payload[1] == 0xff:
		order = func(b []byte) uint16 { return uint16(b[1]) | uint16(b[0])<<8 }
	default:
		return string(payload)
	}
	payload = payload[2:]
	units := make([]uint16, 0, len(payload)/2)
	for ; len(payload) >= 2; payload = payload[2:] {
		units = append(units, order(payload))
	}
	return string(utf16.Decode(units))
}

// splitSetting splits a logical line at the first "=" that is not escaped
// with a backslash, unescaping the key.
func splitSetting(line string) (key, value string, found bool) {
	var name strings.Builder
	for i := 0; i < len(line); i++ {
		switch c := line[i]; {
		case c == '\\' && i+1 < len(line):
			i++
			name.WriteByte(line[i])
		case c == '=':
			return strings.TrimSpace(name.String()), strings.TrimSpace(line[i+1:]), true
		default:
			name.WriteByte(c)
		}
	}
	return "", "", false
}

// parseStanzaName extracts the name from a trimmed line starting with "[".
func parseStanzaName(line string) string {
	name := line[1:]
	if end := strings.LastIndexByte(name, ']'); end >= 0 {
		name = name[:end]
	}
	return name
}

// expandStanzaName replaces $NAME references in a stanza name the way btool
// does: names that resolve are substituted, anything else is left untouched,
// and ${NAME} is not recognized. $SPLUNK_HOME/etc, with either slash style,
// resolves to the environment's etc directory independently of --dir.
func expandStanzaName(name string, vars expansionVars) string {
	if !strings.Contains(name, "$") {
		return name
	}
	var out strings.Builder
	for i := 0; i < len(name); {
		if name[i] != '$' {
			out.WriteByte(name[i])
			i++
			continue
		}
		end := i + 1
		for end < len(name) && isVarChar(name[end]) {
			end++
		}
		varName := name[i+1 : end]
		value, ok := vars.lookup(varName)
		if !ok {
			out.WriteByte('$')
			i++
			continue
		}
		if varName == "SPLUNK_HOME" {
			if rest := name[end:]; len(rest) >= 4 && isStanzaPathSeparator(rest[0]) && rest[1:4] == "etc" && (len(rest) == 4 || isStanzaPathSeparator(rest[4])) {
				value = vars.etcDir
				end += len("/etc")
			}
		}
		out.WriteString(value)
		i = end
	}
	return out.String()
}

func isStanzaPathSeparator(c byte) bool {
	return c == '/' || c == '\\'
}

func isVarChar(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

type expansionVars struct {
	lookupEnv func(string) (string, bool)
	etcDir    string
	home      string
}

func (v expansionVars) lookup(name string) (string, bool) {
	switch name {
	case "":
		return "", false
	case "SPLUNK_HOME":
		return v.home, true
	case "SPLUNK_ETC":
		return v.etcDir, true
	case "SPLUNK_DB":
		if value, ok := v.lookupEnv(name); ok && value != "" {
			return value, true
		}
		return v.home + "/var/lib/splunk", true
	}
	return v.lookupEnv(name)
}
