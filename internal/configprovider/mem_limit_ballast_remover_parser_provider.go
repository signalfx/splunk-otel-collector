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
	"regexp"

	"go.opentelemetry.io/collector/config/configparser"
	"go.opentelemetry.io/collector/service/parserprovider"
)

var ballastKeyRegexp *regexp.Regexp

func init() {
	const expr = "processors::memory_limiter(/\\w+)?::ballast_size_mib"
	ballastKeyRegexp, _ = regexp.Compile(expr)
}

// memLimitBallastRemoverParserProvider implements ParserProvider, wraps a
// ParserProvider, and removes any ballast_size_mib key from the memory_limiter
// processor, if one exists, from the wrapped config. This is to ensure that the
// Collector will still start when support for this key gets removed.
type memLimitBallastRemoverParserProvider struct {
	pp parserprovider.ParserProvider
}

var _ parserprovider.ParserProvider = (*memLimitBallastRemoverParserProvider)(nil)

func NewMemLimitRemoverParserProvider(pp parserprovider.ParserProvider) parserprovider.ParserProvider {
	return &memLimitBallastRemoverParserProvider{pp: pp}
}

func (mpp memLimitBallastRemoverParserProvider) Get() (*configparser.ConfigMap, error) {
	cfgMap, err := mpp.pp.Get()
	if err != nil {
		return nil, fmt.Errorf("memLimitBallastRemoverParserProvider.Get(): %w", err)
	}
	out := configparser.NewParser()
	for _, k := range cfgMap.AllKeys() {
		if isMemLimitBallastKey(k) {
			log.Println("Deprecated memory_limiter processor `ballast_size_mib` key found. Removing from config.")
		} else {
			out.Set(k, cfgMap.Get(k))
		}
	}
	return out, nil
}

func isMemLimitBallastKey(k string) bool {
	return ballastKeyRegexp.MatchString(k)
}
