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

package configprovider

import (
	"fmt"
	"log"
	"strings"

	"go.opentelemetry.io/collector/config/configparser"
	"go.opentelemetry.io/collector/service/parserprovider"
)

const (
	serviceExtensionsPath       = "service::extensions"
	extensionsMemoryBallastPath = "extensions::memory_ballast::size_mib"
	memoryBallast               = "memory_ballast"
)

// ballastParserProvider implements ParserProvider and wraps another ParserProvider,
// adding a memory_ballast extension to the wrapped Otel config if one isn't already
// present.
type ballastParserProvider struct {
	pp      parserprovider.ParserProvider
	sizeMib int
}

var _ parserprovider.ParserProvider = (*ballastParserProvider)(nil)

func BallastParserProvider(pp parserprovider.ParserProvider, sizeMib int) parserprovider.ParserProvider {
	return &ballastParserProvider{
		pp:      pp,
		sizeMib: sizeMib,
	}
}

func (bpp ballastParserProvider) Get() (*configparser.ConfigMap, error) {
	cfgMap, err := bpp.pp.Get()
	if err != nil {
		return nil, fmt.Errorf("ballastParserProvider.Get(): %w", err)
	}
	if hasBallastExtension(cfgMap) {
		return cfgMap, nil
	}
	log.Println("Extension `memory_ballast` not found in config. Adding and enabling a `memory_ballast` extension.")
	return cfgMapWithBallastExt(cfgMap, bpp.sizeMib), nil
}

func hasBallastExtension(cfgMap *configparser.ConfigMap) bool {
	for _, key := range cfgMap.AllKeys() {
		if key == serviceExtensionsPath {
			if exts, ok := cfgMap.Get(key).([]interface{}); ok && extensionsContainMemoryBallast(exts) {
				return true
			}
		}
	}
	return false
}

func extensionsContainMemoryBallast(extensions []interface{}) bool {
	for _, v := range extensions {
		if s, ok := v.(string); ok && isMemoryBallastComponent(s) {
			return true
		}
	}
	return false
}

func isMemoryBallastComponent(extName string) bool {
	return extName == memoryBallast || strings.HasPrefix(extName, memoryBallast+"/")
}

func cfgMapWithBallastExt(parser *configparser.ConfigMap, sizeMib int) *configparser.ConfigMap {
	out := configparser.NewParser()
	for _, k := range parser.AllKeys() {
		out.Set(k, parser.Get(k))
	}
	out.Set(extensionsMemoryBallastPath, sizeMib)
	out.Set(serviceExtensionsPath, []interface{}{memoryBallast})
	return out
}
