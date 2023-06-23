// Copyright Splunk, Inc.
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

package telemetry

import (
	"bytes"
	"crypto/md5" // #nosec this is not for cryptographic purposes
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/signalfx/splunk-otel-collector/tests/internal/version"
)

const (
	anyValue                = "<ANY>"
	buildVersionPlaceholder = "<VERSION_FROM_BUILD>"
)

func marshal(y any) string {
	b := &bytes.Buffer{}
	enc := yaml.NewEncoder(b)
	enc.SetIndent(2)
	if err := enc.Encode(y); err != nil {
		panic(err)
	}
	return b.String()
}

type Resource struct {
	Attributes *map[string]any `yaml:"attributes,omitempty"`
}

func (resource Resource) String() string {
	return marshal(resource)
}

// Hash provides an md5 hash determined by Resource content.
func (resource Resource) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(resource.String()))) // #nosec
}

// Equals determines the equivalence of two Resource items by their Attributes.
func (resource Resource) Equals(toCompare Resource) bool {
	return attributesAreEqual(resource.Attributes, toCompare.Attributes)
}

func (resource Resource) FillDefaultValues() {
	if resource.Attributes != nil {
		for k, v := range *resource.Attributes {
			if v == buildVersionPlaceholder {
				(*resource.Attributes)[k] = version.Version
			}
		}
	}
}

type InstrumentationScope struct {
	Attributes *map[string]any `yaml:"attributes,omitempty"`
	Name       string          `yaml:"name,omitempty"`
	Version    string          `yaml:"version,omitempty"`
}

func (is InstrumentationScope) String() string {
	return marshal(is)
}

// Hash provides an md5 hash determined by InstrumentationScope fields.
func (is InstrumentationScope) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(is.String()))) // #nosec
}

// Equals determines the equivalence of two InstrumentationScope items.
func (is InstrumentationScope) Equals(toCompare InstrumentationScope) bool {
	if is.Name != anyValue {
		if is.Name != toCompare.Name {
			return false
		}
	}
	if is.Version != anyValue {
		if is.Version != toCompare.Version {
			return false
		}
	}
	return attributesAreEqual(is.Attributes, toCompare.Attributes)
}

// attributesAreEqual determines if the provided `attrs` are the same as
// `toCompare`, accounting for <ANY> values in `attrs`.
func attributesAreEqual(attrs, toCompare *map[string]any) bool {
	if attrs == nil {
		return true
	}
	if toCompare == nil {
		return false
	}
	if len(*attrs) != len(*toCompare) {
		return false
	}

	rAttrs := map[string]any{}
	tcAttrs := map[string]any{}

	for k, v := range *attrs {
		tcV, ok := (*toCompare)[k]
		if !ok {
			return false
		}
		if v == anyValue {
			continue
		}
		rAttrs[k] = v
		tcAttrs[k] = tcV
	}

	return reflect.DeepEqual(rAttrs, tcAttrs)
}
