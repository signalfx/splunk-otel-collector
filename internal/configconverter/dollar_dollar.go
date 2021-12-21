// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configconverter

import (
	"log"
	"regexp"

	"go.opentelemetry.io/collector/config"
)

// ReplaceDollarDollar replaces any $${foo:MY_VAR} config source variables with
// ${foo:MY_VAR}. These might exist because of customers working around a bug
// in how the Collector expanded these variables.
func ReplaceDollarDollar(m *config.Map) *config.Map {
	re := dollarDollarRegex()
	for _, k := range m.AllKeys() {
		v := m.Get(k)
		if s, ok := v.(string); ok {
			replaced := replaceDollarDollar(re, s)
			if replaced != s {
				log.Printf("[WARNING] the notation %q is no longer recommended. Please replace with %q.\n", v, replaced)
				m.Set(k, replaced)
			}
		}
	}
	return m
}

func dollarDollarRegex() *regexp.Regexp {
	return regexp.MustCompile(`\$\${(.+?:.+?)}`)
}

func replaceDollarDollar(re *regexp.Regexp, s string) string {
	return re.ReplaceAllString(s, "${$1}")
}
