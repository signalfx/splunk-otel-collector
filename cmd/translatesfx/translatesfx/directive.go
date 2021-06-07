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

package translatesfx

import (
	"fmt"
	"strings"
)

type source int

const (
	directiveSourceUnknown source = iota
	directiveSourceFile
	directiveSourceEnv
	directiveSourceZK
	directiveSourceEtcd2
	directiveSourceConsul
	directiveSourceVault
)

type directive struct {
	fromPath    string
	defaultV    string
	fromType    source
	isDirective bool
	flatten     bool
	optional    bool
}

func parseDirective(m map[interface{}]interface{}) (out directive, err error) {
	fromRaw, ok := m["#from"]
	if !ok {
		return
	}
	out.isDirective = true
	from, ok := fromRaw.(string)
	if !ok {
		return out, fmt.Errorf("error parsing from %v", m)
	}

	out.fromType, out.fromPath, err = parseFrom(from)
	if err != nil {
		return
	}

	out.flatten, err = parseFlatten(m)
	if err != nil {
		return
	}

	out.defaultV, err = parseDefault(m)
	if err != nil {
		return
	}

	out.optional, err = parseOptional(m)
	if err != nil {
		return
	}

	return
}

func parseFrom(from string) (s source, target string, err error) {
	idx := strings.Index(from, ":")
	if idx == -1 {
		return directiveSourceFile, from, nil
	}
	s = strToSource(from[:idx])
	if s == directiveSourceUnknown {
		return s, target, fmt.Errorf("#from fromType type unknown: %s", from)
	}
	target = from[idx+1:]
	return
}

func strToSource(s string) source {
	switch s {
	case "", "file":
		return directiveSourceFile
	case "env":
		return directiveSourceEnv
	case "zookeeper", "zk":
		return directiveSourceZK
	case "etcd2":
		return directiveSourceEtcd2
	case "consul":
		return directiveSourceConsul
	case "vault":
		return directiveSourceVault
	}
	return directiveSourceUnknown
}

func parseFlatten(m map[interface{}]interface{}) (bool, error) {
	return parseField(m, "flatten")
}

func parseOptional(m map[interface{}]interface{}) (bool, error) {
	return parseField(m, "optional")
}

func parseField(m map[interface{}]interface{}, field string) (bool, error) {
	v, ok := m[field]
	if !ok {
		return false, nil
	}
	out, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf(`unable to parse field %q: %v`, field, m)
	}
	return out, nil
}

func parseDefault(m map[interface{}]interface{}) (string, error) {
	rawDefault, ok := m["default"]
	if !ok {
		return "", nil
	}
	out, ok := rawDefault.(string)
	if !ok {
		return "", fmt.Errorf("unable to parse default value: %v", m)
	}
	return out, nil
}
