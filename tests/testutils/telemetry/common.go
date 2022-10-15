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
	"crypto/md5" // #nosec this is not for cryptographic purposes
	"fmt"
	"reflect"

	"gopkg.in/yaml.v2"
)

const buildVersionPlaceholder = "<FROM_BUILD>"

type Resource struct {
	Attributes map[string]any `yaml:"attributes,omitempty"`
}

func (resource Resource) String() string {
	out, err := yaml.Marshal(resource.Attributes)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// Hash provides an md5 hash determined by Resource content.
func (resource Resource) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(resource.String()))) // #nosec
}

// Equals determines the equivalence of two Resource items by their Attributes.
// TODO: ensure that Resource.Hash equivalence is valid given all possible Attribute values.
func (resource Resource) Equals(toCompare Resource) bool {
	// if either attribute map is uninitialized reflection equality is a false negative
	if len(resource.Attributes) == 0 && len(toCompare.Attributes) == 0 {
		return true
	}
	return reflect.DeepEqual(resource.Attributes, toCompare.Attributes)
}

type InstrumentationScope struct {
	Attributes map[string]any `yaml:"attributes,omitempty"`
	Name       string         `yaml:"name,omitempty"`
	Version    string         `yaml:"version,omitempty"`
}

func (is InstrumentationScope) String() string {
	out, err := yaml.Marshal(is)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// Hash provides an md5 hash determined by InstrumentationScope fields.
func (is InstrumentationScope) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(is.String()))) // #nosec
}

// Equals determines the equivalence of two InstrumentationScope items.
// TODO: ensure that Resource.Hash equivalence is valid given all possible Attribute values.
func (is InstrumentationScope) Equals(toCompare InstrumentationScope) bool {
	return is.Name == toCompare.Name && is.Version == toCompare.Version
}
func (is InstrumentationScope) Matches(toCompare InstrumentationScope, strict bool) bool {
	if is.Name != toCompare.Name && (strict || is.Name != "") {
		return false
	}
	if is.Version != toCompare.Version && (strict || is.Version != "") {
		return false
	}
	return true
}
