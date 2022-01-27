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

	"go.opentelemetry.io/collector/config"
)

func MoveHecTLS(_ context.Context, in *config.Map) error {
	if in == nil {
		return fmt.Errorf("cannot MoveHecTLS on nil *config.Map")
	}

	const expression = "exporters::splunk_hec(/\\w+)?::(insecure_skip_verify|ca_file|cert_file|key_file)"
	re, _ := regexp.Compile(expression)
	out := config.NewMap()
	unsupportedKeyFound := false
	for _, k := range in.AllKeys() {
		v := in.Get(k)
		match := re.FindStringSubmatch(k)
		if match == nil {
			out.Set(k, v)
		} else {
			tlsKey := fmt.Sprintf("exporters::splunk_hec%s::tls::%s", match[1], match[2])
			log.Printf("Unsupported key found: %s. Moving to %s\n", k, tlsKey)
			out.Set(tlsKey, v)
			unsupportedKeyFound = true
		}
	}
	if unsupportedKeyFound {
		log.Println(
			"[WARNING] `exporters` -> `splunk_hec` -> `insecure_skip_verify|ca_file|cert_file|key_file` " +
				"parameters have moved under `tls`. Please update your config. " +
				"https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/5433",
		)
	}

	*in = *out
	return nil
}
