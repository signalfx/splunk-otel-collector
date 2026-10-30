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
	"strings"
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
//   - A stanza header is everything between the first "[" and the last "]".
func parseConf(payload []byte) []setting {
	text := strings.TrimPrefix(string(payload), "\xef\xbb\xbf")
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
		for strings.HasSuffix(line, "\\") {
			line = strings.TrimSuffix(line, "\\")
			if i+1 >= len(lines) {
				break
			}
			i++
			line += "\n" + lines[i]
		}
		key, value, found := splitSetting(line)
		if !found {
			continue
		}
		settings = append(settings, setting{stanza: stanza, key: key, value: value})
	}
	return settings
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
// and ${NAME} is not recognized. $SPLUNK_HOME/etc resolves to the etc
// directory in use.
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
			if rest := name[end:]; strings.HasPrefix(rest, "/etc") && (len(rest) == len("/etc") || rest[len("/etc")] == '/') {
				value = vars.etcDir
				end += len("/etc")
			}
		}
		out.WriteString(value)
		i = end
	}
	return out.String()
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
