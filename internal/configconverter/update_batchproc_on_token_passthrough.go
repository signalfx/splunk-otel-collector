// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"go.opentelemetry.io/collector/confmap"
)

func UpdateBatchProcOnTokenPassthrough(_ context.Context, in *confmap.Conf) error {
	if in == nil {
		return nil
	}

	out := map[string]any{}

	// Check for sapm receiver with include_metadata set to true
	if in.IsSet("receivers::sapm::include_metadata") {
		if accessTokenPassthrough, ok := in.Get("receivers::sapm::include_metadata").(bool); ok && accessTokenPassthrough {
			// Add metadata_keys to batch processor
			switch batchProcessor := in.Get("processors::batch").(type) {
			case nil:
				out["processors::batch"] = map[string]any{
					"metadata_keys": []interface{}{"X-SF-Token"},
				}
			case map[string]interface{}:
				batchProcessor["metadata_keys"] = []any{"X-SF-Token"}
				out["processors::batch"] = batchProcessor
			default:
				return fmt.Errorf("unexpected type for processors::batch: %T", batchProcessor)
			}
		}
	}

	// Merge the modified configuration back into the original Conf
	modifiedConf := confmap.NewFromStringMap(out)
	return in.Merge(modifiedConf)
}
