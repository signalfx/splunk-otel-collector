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
	"context"
	"fmt"
	"log"
	"regexp"

	"go.opentelemetry.io/collector/confmap"
)

type MoveOTLPInsecureKey struct{}

func (MoveOTLPInsecureKey) Convert(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return fmt.Errorf("cannot MoveOTLPInsecureKey on nil *confmap.Conf")
	}

	const expr = "exporters::otlp(/\\w+)?::insecure"
	insecureRE := regexp.MustCompile(expr)
	out := map[string]interface{}{}
	var deprecatedOTLPConfigFound bool
	for _, k := range in.AllKeys() {
		v := in.Get(k)
		match := insecureRE.FindStringSubmatch(k)
		if match == nil {
			out[k] = v
		} else {
			tlsKey := fmt.Sprintf("exporters::otlp%s::tls::insecure", match[1])
			log.Printf("Unsupported key found: %s. Moving to %s\n", k, tlsKey)
			out[tlsKey] = v
			deprecatedOTLPConfigFound = true
		}
	}
	if deprecatedOTLPConfigFound {
		log.Println("[WARNING] `exporters` -> `otlp` -> `insecure` parameter is " +
			"deprecated. Please update the config according to the guideline: " +
			"https://github.com/signalfx/splunk-otel-collector#from-0350-to-0360.")
	}

	*in = *confmap.NewFromStringMap(out)
	return nil
}
